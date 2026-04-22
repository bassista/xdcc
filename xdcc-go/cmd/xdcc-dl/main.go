package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"xdcc-go/internal/downloader"
	"xdcc-go/internal/entities"
)

func main() {
	var (
		server           string
		out              string
		throttle         string
		connectTimeout   int
		stallTimeout     int
		fallbackChannel  string
		waitTime         int
		username         string
		channelJoinDelay int
		verbosity        int
		quiet            bool
	)

	cmd := &cobra.Command{
		Use:   "xdcc-dl <message>",
		Short: "Download a file via XDCC IRC protocol",
		Long: `xdcc-dl downloads files from IRC bots using the XDCC protocol.
The message should be in the format: /msg <bot> xdcc send #<pack>
Supports ranges (#1-10), steps (#1-10;2), and comma lists (#1,2,3).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := args[0]

			packs, err := entities.ParseXDCCMessage(message, ".", server)
			if err != nil {
				return fmt.Errorf("invalid XDCC message: %w", err)
			}
			if len(packs) == 0 {
				return fmt.Errorf("no packs found in message: %s", message)
			}

			entities.PreparePacks(packs, out)

			throttleBytes, err := entities.ParseThrottle(throttle)
			if err != nil {
				return fmt.Errorf("invalid throttle value %q: %w", throttle, err)
			}

			downloader.DownloadPacks(packs, downloader.Options{
				ConnectTimeout:   connectTimeout,
				StallTimeout:     stallTimeout,
				FallbackChannel:  fallbackChannel,
				ThrottleBytes:    throttleBytes,
				WaitTime:         waitTime,
				Username:         username,
				ChannelJoinDelay: channelJoinDelay,
				Verbosity:        verbosityLevel(verbosity, quiet),
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", "irc.rizon.net",
		"IRC server address")
	cmd.Flags().StringVarP(&out, "out", "o", "",
		"Output file path (defaults to pack filename)")
	cmd.Flags().StringVarP(&throttle, "throttle", "t", "-1",
		"Download speed limit (e.g. 50M, 100K, 1G). -1 = unlimited")
	cmd.Flags().IntVar(&connectTimeout, "connect-timeout", 120,
		"Seconds to wait for the bot to initiate DCC (before the transfer starts)")
	cmd.Flags().IntVar(&stallTimeout, "stall-timeout", 60,
		"Seconds of no transfer progress before aborting (0 = disabled)")
	cmd.Flags().StringVar(&fallbackChannel, "fallback-channel", "",
		"IRC channel to join if WHOIS finds none")
	cmd.Flags().IntVar(&waitTime, "wait-time", 0,
		"Seconds to wait before sending the XDCC request")
	cmd.Flags().StringVar(&username, "username", "",
		"IRC nickname to use (random if not set)")
	cmd.Flags().IntVar(&channelJoinDelay, "channel-join-delay", -1,
		"Seconds to wait after connecting before joining channels (-1 = random 5-10s)")
	cmd.Flags().CountVarP(&verbosity, "verbose", "v", "Verbose output (-v=verbose, -vv=debug)")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (suppress all output)")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// verbosityLevel converts count+quiet flags to verbosity int.
// quiet => -1, default => 0, -v => 1, -vv => 2
func verbosityLevel(count int, quiet bool) int {
	if quiet {
		return -1
	}
	return count
}
