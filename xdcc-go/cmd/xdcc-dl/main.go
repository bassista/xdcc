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

The message must be in the format:  /msg <bot> xdcc send #<pack>
Pack number supports ranges, steps, and comma lists:
  /msg MyBot xdcc send #5        single pack
  /msg MyBot xdcc send #1-10     packs 1 through 10
  /msg MyBot xdcc send #1-10;2   packs 1,3,5,7,9 (every 2nd)
  /msg MyBot xdcc send #1,3,7    specific packs

The IRC server is detected automatically from the bot name prefix when
possible. Use --server to override with an explicit address.

Verbosity levels:
  (default)  show connection and download progress
  -v         also show bot notices, channel joins, WHOIS results
  -vv        full debug (DNS, DCC internals, all IRC events)
  -q         suppress all output`,
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
		"IRC server address (host or host:port). Overrides automatic server detection from bot name")
	cmd.Flags().StringVarP(&out, "out", "o", "",
		"Output directory or file path (defaults to current directory with pack filename)")
	cmd.Flags().StringVarP(&throttle, "throttle", "t", "-1",
		"Download speed limit in bytes/s (e.g. 512K, 2M, 1G). -1 = unlimited")
	cmd.Flags().IntVar(&connectTimeout, "connect-timeout", 120,
		"Seconds to wait for the bot to initiate the DCC transfer (default: 120)")
	cmd.Flags().IntVar(&stallTimeout, "stall-timeout", 60,
		"Seconds of no transfer progress before aborting. 0 = disabled (default: 60)")
	cmd.Flags().StringVar(&fallbackChannel, "fallback-channel", "",
		"IRC channel to join if WHOIS returns no channels for the bot")
	cmd.Flags().IntVar(&waitTime, "wait-time", 0,
		"Extra seconds to wait before sending the XDCC request (default: 0)")
	cmd.Flags().StringVar(&username, "username", "",
		"IRC nickname to use. A random suffix is appended automatically. Default: random")
	cmd.Flags().IntVar(&channelJoinDelay, "channel-join-delay", -1,
		"Seconds to wait after connecting before sending WHOIS. -1 = random 5-10s (default: -1)")
	cmd.Flags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity: -v shows bot notices, -vv shows full debug info")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output including progress")

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
