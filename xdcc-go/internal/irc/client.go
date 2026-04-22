// Package irc implements the XDCC IRC client using the girc library.
// It handles connecting to an IRC server, performing WHOIS to find the
// bot's channel, joining that channel, sending the XDCC request, and
// managing the resulting DCC file transfer.
package irc

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lrstanley/girc"
	"xdcc-go/internal/entities"
)

// XDCCDownloadError represents a typed error from the XDCC download process.
type XDCCDownloadError struct {
	Kind    string
	Message string
}

func (e *XDCCDownloadError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

var (
	ErrTimeout           = &XDCCDownloadError{Kind: "timeout", Message: "download timed out"}
	ErrBotNotFound       = &XDCCDownloadError{Kind: "bot_not_found", Message: "bot does not exist on server"}
	ErrPackAlreadyReq    = &XDCCDownloadError{Kind: "pack_already_requested", Message: "pack already requested, try again later"}
	ErrAlreadyDownloaded = &XDCCDownloadError{Kind: "already_downloaded", Message: "file already downloaded"}
	ErrBotDenied         = &XDCCDownloadError{Kind: "bot_denied", Message: "bot denied the XDCC request"}
	ErrUnrecoverable     = &XDCCDownloadError{Kind: "unrecoverable", Message: "unrecoverable error (IP banned?)"}
	ErrDownloadFailed    = &XDCCDownloadError{Kind: "download_failed", Message: "download did not complete"}
)

// DownloadOptions configures a download session.
type DownloadOptions struct {
	ConnectTimeout   int    // seconds to wait for bot to initiate DCC (default 120)
	StallTimeout     int    // seconds of no transfer progress before aborting (0 = disabled, default 60)
	FallbackChannel  string // used if WHOIS finds no channels
	ThrottleBytes    int64  // bytes/sec limit, -1 = unlimited
	WaitTime         int    // seconds to wait before sending XDCC request
	Username         string // IRC nick to use; empty = random
	ChannelJoinDelay int    // seconds to wait after connecting before joining channels; -1 = random 5-10
}

// Client manages a single XDCC pack download.
type Client struct {
	pack    *entities.XDCCPack
	opts    DownloadOptions
	irc     *girc.Client
	verbose bool

	// DCC state
	mu           sync.Mutex
	peerAddr     string // stored on DCC SEND, used again on DCC ACCEPT
	dccConn      net.Conn
	dccFile      *os.File
	progress     int64
	filesize     int64
	dccTimestamp time.Time
	downloading  bool

	// ACK queue
	ackQueue chan []byte

	// flow control
	messageSent     atomic.Bool
	downloadDone    chan struct{}
	downloadStarted chan struct{} // closed when DCC TCP connection is established
	downloadError   error
	connectTime     time.Time
	closeOnce       sync.Once // ensures downloadDone and ackQueue are closed only once
	startOnce       sync.Once // ensures downloadStarted is closed only once

	// last bot notice message (for error reporting)
	lastBotNotice string

	// stall detection: unix nanoseconds of last received byte
	lastActivity atomic.Int64

	// timing
	downStartTime time.Time
}

// NewClient creates a new XDCC download client.
func NewClient(pack *entities.XDCCPack, opts DownloadOptions, verbose bool) *Client {
	if opts.ChannelJoinDelay < 0 {
		// random 5-10 seconds
		n, _ := rand.Int(rand.Reader, big.NewInt(6))
		opts.ChannelJoinDelay = int(n.Int64()) + 5
	}
	if opts.ConnectTimeout <= 0 {
		opts.ConnectTimeout = 120
	}
	if opts.StallTimeout < 0 {
		opts.StallTimeout = 0
	}
	return &Client{
		pack:            pack,
		opts:            opts,
		verbose:         verbose,
		downloadDone:    make(chan struct{}),
		downloadStarted: make(chan struct{}),
		ackQueue:        make(chan []byte, 256),
	}
}

// randomUsername generates a random IRC nickname.
func randomUsername() string {
	firstNames := []string{"Alice", "Bob", "Charlie", "Dave", "Eve", "Frank", "Grace", "Hank"}
	lastNames := []string{"Smith", "Jones", "Brown", "Wilson", "Taylor", "Davis", "Clark", "Lewis"}
	n1, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstNames))))
	n2, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lastNames))))
	num, _ := rand.Int(rand.Reader, big.NewInt(90))
	suffix := randomSuffix(3)
	return fmt.Sprintf("%s%s%d%s",
		firstNames[n1.Int64()],
		lastNames[n2.Int64()],
		num.Int64()+10,
		suffix)
}

func randomSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

// Download connects to IRC and downloads the pack.
// Returns the path to the downloaded file, or an error.
func (c *Client) Download() (string, error) {
	nick := c.opts.Username
	if nick == "" {
		nick = randomUsername()
	} else {
		nick = nick + randomSuffix(3)
	}

	c.infof("Connecting to %s:%d as '%s'", c.pack.Server.Address, c.pack.Server.Port, nick)
	c.logf("Pack: #%d from bot '%s'", c.pack.PackNumber, c.pack.Bot)

	c.irc = girc.New(girc.Config{
		Server:      c.pack.Server.Address,
		Port:        c.pack.Server.Port,
		Nick:        nick,
		User:        nick,
		Name:        nick,
		PingDelay:   30 * time.Second,
		PingTimeout: 60 * time.Second,
	})

	c.registerHandlers()

	// Run IRC event loop in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.irc.Connect()
	}()

	// Phase 1: wait for the DCC transfer to start (connect timeout covers
	// IRC connect + channel join + bot response time).
	connectTimeout := time.Duration(c.opts.ConnectTimeout+c.opts.WaitTime+c.opts.ChannelJoinDelay+30) * time.Second
	c.logf("Connect timeout: %s", connectTimeout)
	select {
	case <-c.downloadStarted:
		// Transfer started — move to phase 2
		c.logf("Transfer started, switching to stall detection")
	case <-c.downloadDone:
		// Completed or errored before transfer started
		goto done
	case err := <-errCh:
		if err != nil && c.downloadError == nil {
			c.downloadError = err
		}
		goto done
	case <-time.After(connectTimeout):
		c.irc.Close()
		if c.downloadError == nil {
			c.infof("Connect timeout (%s): bot did not initiate DCC transfer in time", connectTimeout)
			c.downloadError = ErrTimeout
		}
		goto done
	}

	// Phase 2: transfer is in progress.
	// Stall detection is handled by stallWatcher() goroutine (if StallTimeout > 0).
	// Wait for completion or IRC error.
	if c.opts.StallTimeout > 0 {
		go c.stallWatcher()
	}
	select {
	case <-c.downloadDone:
		// normal completion or error during transfer
	case err := <-errCh:
		if err != nil && c.downloadError == nil {
			c.downloadError = err
		}
	}

done:
	if c.downloadError != nil {
		return "", c.downloadError
	}
	return c.pack.GetFilepath(), nil
}

func (c *Client) registerHandlers() {
	// After successful connection
	c.irc.Handlers.Add(girc.CONNECTED, func(client *girc.Client, e girc.Event) {
		c.connectTime = time.Now()
		c.infof("Connected to server")
		c.logf("Waiting %ds before WHOIS (channel join delay)", c.opts.ChannelJoinDelay)
		time.Sleep(time.Duration(c.opts.ChannelJoinDelay) * time.Second)
		c.logf("Sending WHOIS for bot '%s'", c.pack.Bot)
		// WHOIS the bot to find its channels
		client.Cmd.Whois(c.pack.Bot)
		// Start timeout watcher
		go c.timeoutWatcher(client)
	})

	// End of WHOIS - if no channels found, try fallback or send anyway
	c.irc.Handlers.Add(girc.RPL_ENDOFWHOIS, func(client *girc.Client, e girc.Event) {
		c.logf("End of WHOIS")
		if !c.messageSent.Load() {
			if c.opts.FallbackChannel != "" {
				ch := c.opts.FallbackChannel
				if !strings.HasPrefix(ch, "#") {
					ch = "#" + ch
				}
				c.logf("No channels from WHOIS; joining fallback channel %s", ch)
				client.Cmd.Join(ch)
			} else {
				c.logf("No channels from WHOIS and no fallback; sending XDCC request directly")
				c.sendXDCCRequest(client)
			}
		}
	})

	// WHOIS channels info - join each channel the bot is in
	c.irc.Handlers.Add(girc.RPL_WHOISCHANNELS, func(client *girc.Client, e girc.Event) {
		if len(e.Params) < 2 {
			return
		}
		c.logf("WHOIS channels: %s", e.Params[len(e.Params)-1])
		rawChannels := e.Params[len(e.Params)-1]
		parts := strings.Fields(rawChannels)
		for _, part := range parts {
			part = strings.TrimLeft(part, "@+%&~")
			if strings.HasPrefix(part, "#") {
				c.logf("Joining channel %s", part)
				time.Sleep(time.Duration(1+randN(2)) * time.Second)
				client.Cmd.Join(part)
			}
		}
	})

	// On JOIN - send XDCC request after joining
	c.irc.Handlers.Add(girc.JOIN, func(client *girc.Client, e girc.Event) {
		// Only act if we were the one joining
		if e.Source == nil || !strings.EqualFold(e.Source.Name, client.GetNick()) {
			return
		}
		c.logf("Joined channel: %s", e.Params[0])
		if !c.messageSent.Load() {
			c.sendXDCCRequest(client)
		}
	})

	// CTCP DCC handler (DCC SEND / DCC ACCEPT for resume)
	c.irc.CTCP.Set("DCC", func(client *girc.Client, ctcp girc.CTCPEvent) {
		c.handleDCC(client, ctcp.Text)
	})

	// Handle NOTICE from bot (pack already requested, slots busy, denied, etc.)
	c.irc.Handlers.Add(girc.NOTICE, func(client *girc.Client, e girc.Event) {
		notice := e.Last()
		msg := strings.ToLower(notice)
		// Always show NOTICE messages from the bot
		c.infof("Bot notice: %s", notice)

		alreadyReqMsgs := []string{"you already requested", "richiesto questo pack!"}
		blockedMsgs := []string{"xdcc send negato", "numero pack errato", "invalid pack number", "gli slots sono occupati", "denied"}

		for _, s := range alreadyReqMsgs {
			if strings.Contains(msg, s) {
				c.mu.Lock()
				c.lastBotNotice = notice
				c.mu.Unlock()
				c.finishWithError(ErrPackAlreadyReq)
				return
			}
		}
		for _, s := range blockedMsgs {
			if strings.Contains(msg, s) {
				c.mu.Lock()
				c.lastBotNotice = notice
				c.mu.Unlock()
				c.finishWithError(ErrBotDenied)
				return
			}
		}
	})

	// Bot not found on server
	c.irc.Handlers.Add(girc.ERR_NOSUCHNICK, func(client *girc.Client, e girc.Event) {
		c.infof("Bot '%s' not found on server", c.pack.Bot)
		c.finishWithError(ErrBotNotFound)
	})

	// IRC error (ban, etc.)
	c.irc.Handlers.Add(girc.ERROR, func(client *girc.Client, e girc.Event) {
		c.infof("IRC error: %s", e.Last())
		c.finishWithError(ErrUnrecoverable)
	})
}

func (c *Client) sendXDCCRequest(client *girc.Client) {
	if c.messageSent.Swap(true) {
		return // already sent
	}
	if c.opts.WaitTime > 0 {
		c.logf("Waiting %ds before sending XDCC request", c.opts.WaitTime)
		time.Sleep(time.Duration(c.opts.WaitTime) * time.Second)
	}
	msg := c.pack.GetRequestMessage(false)
	c.infof("Sending XDCC request: /msg %s %s", c.pack.Bot, msg)
	client.Cmd.Message(c.pack.Bot, msg)
}

func (c *Client) handleDCC(client *girc.Client, text string) {
	// text is like: SEND filename ip port filesize
	//            or: ACCEPT filename port position
	parts := splitDCC(text)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToUpper(parts[0])
	switch cmd {
	case "SEND":
		c.handleDCCSend(client, parts)
	case "ACCEPT":
		c.handleDCCAccept(parts)
	default:
		c.logf("Unknown DCC command: %s", cmd)
	}
}

func (c *Client) handleDCCSend(client *girc.Client, parts []string) {
	if len(parts) < 5 {
		c.logf("Malformed DCC SEND: %v", parts)
		return
	}
	filename := parts[1]
	ipNum := parts[2]
	port := parts[3]
	sizeStr := parts[4]

	peerAddr := ipNumToQuad(ipNum) + ":" + port
	filesize := parseI64(sizeStr)

	c.pack.SetFilename(filename, false)
	c.filesize = filesize

	// Store peer address for potential DCC RESUME → ACCEPT flow
	c.mu.Lock()
	c.peerAddr = peerAddr
	c.mu.Unlock()

	c.infof("DCC SEND: file=%s addr=%s size=%s", filename, peerAddr, entities.HumanReadableBytes(filesize))

	// Check for resume
	existingPath := c.pack.GetFilepath()
	c.logf("Checking for existing file at: %s", existingPath)
	if fi, err := os.Stat(existingPath); err == nil {
		pos := fi.Size()
		c.logf("Existing file size: %s, DCC file size: %s", entities.HumanReadableBytes(pos), entities.HumanReadableBytes(filesize))
		if pos >= filesize {
			c.infof("File already fully downloaded (local: %s >= remote: %s), skipping",
				entities.HumanReadableBytes(pos), entities.HumanReadableBytes(filesize))
			c.finishWithError(ErrAlreadyDownloaded)
			return
		}
		// Request resume
		c.progress = pos
		resumeParam := fmt.Sprintf("\"%s\" %s %d", filename, port, pos)
		c.infof("Resuming download from %s / %s", entities.HumanReadableBytes(pos), entities.HumanReadableBytes(filesize))
		c.logf("Sending DCC RESUME: %s", resumeParam)
		// Send CTCP DCC RESUME to the bot
		client.Cmd.SendCTCP(c.pack.Bot, "DCC", "RESUME "+resumeParam)
		return
	}

	// Start fresh download
	c.startDownload(peerAddr, false)
}

func (c *Client) handleDCCAccept(parts []string) {
	if len(parts) < 4 {
		return
	}
	port := parts[2]
	// position := parts[3] // already set in progress

	// Find peer address from the previous SEND (reuse it)
	// We need to reconstruct - store it from SEND
	// For simplicity, we re-use the stored address
	c.logf("DCC ACCEPT: resuming download")
	_ = port
	// Start appending
	c.startDownloadAppend()
}

// We need to store the peer address between SEND and ACCEPT
// Add a field to Client for this purpose
func (c *Client) startDownload(addr string, appendMode bool) {
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if appendMode {
		flag = os.O_APPEND | os.O_WRONLY
	}

	path := c.pack.GetFilepath()
	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		c.logf("Cannot open file: %v", err)
		c.finishWithError(fmt.Errorf("cannot open file: %w", err))
		return
	}

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		f.Close()
		c.logf("DCC connection failed: %v", err)
		c.finishWithError(fmt.Errorf("DCC connection failed: %w", err))
		return
	}

	c.mu.Lock()
	c.dccFile = f
	c.dccConn = conn
	c.downStartTime = time.Now()
	c.dccTimestamp = time.Now()
	c.downloading = true
	c.mu.Unlock()

	c.logf("Starting download (append=%v) to %s", appendMode, path)
	c.infof("Downloading %s → %s", entities.HumanReadableBytes(c.filesize), path)

	// Signal that transfer has started
	c.startOnce.Do(func() { close(c.downloadStarted) })
	c.lastActivity.Store(time.Now().UnixNano())

	go c.ackSender()
	go c.progressPrinter()
	go c.receiveData()
}

func (c *Client) startDownloadAppend() {
	// Called on DCC ACCEPT; peer address was stored earlier
	c.mu.Lock()
	peerAddr := c.peerAddr
	c.mu.Unlock()
	if peerAddr == "" {
		c.finishWithError(ErrDownloadFailed)
		return
	}
	c.startDownload(peerAddr, true)
}

func (c *Client) receiveData() {
	defer func() {
		c.mu.Lock()
		c.downloading = false
		if c.dccFile != nil {
			c.dccFile.Close()
		}
		c.mu.Unlock()

		if c.progress >= c.filesize {
			c.logf("Download complete")
			c.finishSuccess()
		} else {
			c.logf("Download incomplete: got %d of %d bytes", c.progress, c.filesize)
			c.finishWithError(ErrDownloadFailed)
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, err := c.dccConn.Read(buf)
		if n > 0 {
			c.mu.Lock()
			_, werr := c.dccFile.Write(buf[:n])
			c.mu.Unlock()
			if werr != nil {
				c.logf("Write error: %v", werr)
				return
			}
			atomic.AddInt64(&c.progress, int64(n))
			c.lastActivity.Store(time.Now().UnixNano())

			// Throttle
			if c.opts.ThrottleBytes > 0 {
				c.mu.Lock()
				delta := time.Since(c.dccTimestamp).Seconds()
				chunkTime := float64(n) / float64(c.opts.ThrottleBytes)
				sleepTime := chunkTime - delta
				c.dccTimestamp = time.Now()
				c.mu.Unlock()
				if sleepTime > 0 {
					time.Sleep(time.Duration(sleepTime * float64(time.Second)))
				}
			}

			// Enqueue ACK
			c.enqueueACK()
		}
		if err != nil {
			return
		}
	}
}

// ackSender sends ACKs from the queue.
func (c *Client) ackSender() {
	for ack := range c.ackQueue {
		c.mu.Lock()
		conn := c.dccConn
		c.mu.Unlock()
		if conn == nil {
			continue
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.Write(ack)
	}
}

// enqueueACK encodes the current progress as a big-endian uint32 (or uint64) ACK.
func (c *Client) enqueueACK() {
	prog := atomic.LoadInt64(&c.progress)
	var ack []byte
	if prog <= 0xFFFFFFFF {
		ack = make([]byte, 4)
		binary.BigEndian.PutUint32(ack, uint32(prog))
	} else {
		ack = make([]byte, 8)
		binary.BigEndian.PutUint64(ack, uint64(prog))
	}
	select {
	case c.ackQueue <- ack:
	default:
		// drop if queue full
	}
}

// progressPrinter prints download progress periodically.
func (c *Client) progressPrinter() {
	c.mu.Lock()
	for !c.downloading {
		c.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		c.mu.Lock()
	}
	c.mu.Unlock()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastProgress int64
	lastTime := time.Now()

	for {
		select {
		case <-ticker.C:
			prog := atomic.LoadInt64(&c.progress)
			total := c.filesize
			elapsed := time.Since(lastTime).Seconds()
			speed := float64(prog-lastProgress) / elapsed
			lastProgress = prog
			lastTime = time.Now()

			pct := 0.0
			if total > 0 {
				pct = float64(prog) / float64(total) * 100
			}
			fmt.Printf("\r  %.1f%% [%s / %s] %.1f KB/s    ",
				pct,
				entities.HumanReadableBytes(prog),
				entities.HumanReadableBytes(total),
				speed/1024)

			c.mu.Lock()
			dl := c.downloading
			c.mu.Unlock()
			if !dl {
				fmt.Println()
				return
			}
		case <-c.downloadDone:
			fmt.Println()
			return
		}
	}
}

// timeoutWatcher watches for timeout if the XDCC message was never sent.
func (c *Client) timeoutWatcher(client *girc.Client) {
	for !c.messageSent.Load() {
		time.Sleep(1 * time.Second)
		if time.Since(c.connectTime) > time.Duration(c.opts.ConnectTimeout+c.opts.WaitTime)*time.Second {
			c.logf("Timeout: XDCC request never sent")
			c.finishWithError(ErrTimeout)
			client.Close()
			return
		}
	}
}

// stallWatcher monitors transfer progress and closes the connection if no bytes
// are received for StallTimeout seconds. Runs only during an active transfer.
func (c *Client) stallWatcher() {
	stall := time.Duration(c.opts.StallTimeout) * time.Second
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.downloadDone:
			return
		case <-ticker.C:
			last := c.lastActivity.Load()
			if last == 0 {
				continue
			}
			idle := time.Since(time.Unix(0, last))
			c.logf("Stall check: idle for %s (limit %s)", idle.Round(time.Second), stall)
			if idle >= stall {
				c.infof("Transfer stalled for %s (no data received), aborting", idle.Round(time.Second))
				c.finishWithError(ErrTimeout)
				c.irc.Close()
				return
			}
		}
	}
}

func (c *Client) finishSuccess() {
	elapsed := time.Since(c.connectTime)
	speed := float64(c.filesize) / elapsed.Seconds() / 1024
	fmt.Printf("File %s downloaded successfully in %s at %.1f KB/s\n",
		c.pack.Filename,
		formatDuration(elapsed),
		speed)
	c.irc.Close()
	c.closeOnce.Do(func() {
		close(c.downloadDone)
		close(c.ackQueue)
	})
}

func (c *Client) finishWithError(err error) {
	c.mu.Lock()
	if c.downloadError == nil {
		c.downloadError = err
	}
	c.mu.Unlock()
	c.irc.Close()
	c.closeOnce.Do(func() {
		close(c.downloadDone)
		close(c.ackQueue)
	})
}

// LastBotNotice returns the last NOTICE message received from the bot.
func (c *Client) LastBotNotice() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastBotNotice
}

func (c *Client) logf(format string, args ...interface{}) {
	if c.verbose {
		log.Printf("[xdcc] "+format, args...)
	}
}

func (c *Client) infof(format string, args ...interface{}) {
	log.Printf("[xdcc] "+format, args...)
}

// formatDuration formats a duration as "Xs" or "X.Xm".
func formatDuration(d time.Duration) string {
	if d < 90*time.Second {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// ipNumToQuad converts a numeric IP string to dotted-quad notation.
func ipNumToQuad(ipNum string) string {
	n := parseU32(ipNum)
	return fmt.Sprintf("%d.%d.%d.%d",
		(n>>24)&0xFF, (n>>16)&0xFF, (n>>8)&0xFF, n&0xFF)
}

func parseI64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func parseU32(s string) uint32 {
	var v uint32
	fmt.Sscanf(s, "%d", &v)
	return v
}

func randN(n int) int {
	r, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(r.Int64())
}

// splitDCC splits a DCC message text, respecting quoted filenames.
func splitDCC(s string) []string {
	var parts []string
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		if s[0] == '"' {
			end := strings.Index(s[1:], "\"")
			if end < 0 {
				parts = append(parts, s[1:])
				break
			}
			parts = append(parts, s[1:end+1])
			s = strings.TrimSpace(s[end+2:])
		} else {
			sp := strings.IndexByte(s, ' ')
			if sp < 0 {
				parts = append(parts, s)
				break
			}
			parts = append(parts, s[:sp])
			s = strings.TrimSpace(s[sp+1:])
		}
	}
	return parts
}
