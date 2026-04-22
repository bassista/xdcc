// Package downloader orchestrates downloading multiple XDCC packs sequentially.
package downloader

import (
	"errors"
	"fmt"
	"time"

	"xdcc-go/internal/entities"
	xdccirc "xdcc-go/internal/irc"
)

// Options configures the download session.
type Options struct {
	ConnectTimeout   int
	StallTimeout     int
	FallbackChannel  string
	ThrottleBytes    int64
	WaitTime         int
	Username         string
	ChannelJoinDelay int
	Verbose          bool
}

// DownloadPacks downloads a list of XDCC packs sequentially.
// On "pack already requested" it waits 60 seconds before retrying.
func DownloadPacks(packs []*entities.XDCCPack, opts Options) {
	for _, pack := range packs {
		downloadPack(pack, opts, 0)
	}
}

func downloadPack(pack *entities.XDCCPack, opts Options, retryCount int) {
	if retryCount > 3 {
		fmt.Printf("Giving up on pack %d after 3 retries\n", pack.PackNumber)
		return
	}

	ircOpts := xdccirc.DownloadOptions{
		ConnectTimeout:   opts.ConnectTimeout,
		StallTimeout:     opts.StallTimeout,
		FallbackChannel:  opts.FallbackChannel,
		ThrottleBytes:    opts.ThrottleBytes,
		WaitTime:         opts.WaitTime,
		Username:         opts.Username,
		ChannelJoinDelay: opts.ChannelJoinDelay,
	}

	client := xdccirc.NewClient(pack, ircOpts, opts.Verbose)
	_, err := client.Download()
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, xdccirc.ErrAlreadyDownloaded):
		fmt.Printf("File already downloaded (skipping): %s\n", pack.Filename)
	case errors.Is(err, xdccirc.ErrBotDenied):
		notice := client.LastBotNotice()
		if notice != "" {
			fmt.Printf("Bot denied XDCC request: %s\n", notice)
		} else {
			fmt.Printf("Bot denied XDCC request for: %s\n", pack.Filename)
		}
	case errors.Is(err, xdccirc.ErrPackAlreadyReq):
		fmt.Println("Pack already requested. Waiting 60 seconds before retrying...")
		time.Sleep(60 * time.Second)
		downloadPack(pack, opts, retryCount+1)
	case errors.Is(err, xdccirc.ErrBotNotFound):
		fmt.Printf("Bot %s not found on server %s\n", pack.Bot, pack.Server.Address)
	case errors.Is(err, xdccirc.ErrServerUnreachable):
		// Don't retry — the server is down/blocked
		fmt.Printf("Server unreachable (%s): %v\n", pack.Server.Address, err)
		fmt.Println("Tip: use --server to override the IRC server address.")
	case errors.Is(err, xdccirc.ErrTimeout):
		if retryCount < 3 {
			fmt.Println("Retrying...")
			downloadPack(pack, opts, retryCount+1)
		} else {
			fmt.Printf("Giving up on pack %d after timeout\n", pack.PackNumber)
		}
	case errors.Is(err, xdccirc.ErrUnrecoverable):
		fmt.Println("Unrecoverable error (IP banned?). Aborting.")
	case errors.Is(err, xdccirc.ErrDownloadFailed):
		fmt.Printf("Download of %s failed. Retrying...\n", pack.Filename)
		if retryCount < 3 {
			downloadPack(pack, opts, retryCount+1)
		}
	default:
		fmt.Printf("Error downloading pack %d: %v\n", pack.PackNumber, err)
	}
}
