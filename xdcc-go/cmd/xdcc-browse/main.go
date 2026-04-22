package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"xdcc-go/internal/downloader"
	"xdcc-go/internal/entities"
	"xdcc-go/internal/search"
)

func main() {
	var (
		engineName       string
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
		Use:   "xdcc-browse <search_term>",
		Short: "Search for XDCC packs and download interactively",
		Long: `xdcc-browse searches for XDCC packs, displays an interactive selection
menu, and then downloads the chosen pack(s).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			term := args[0]

			engine := search.EngineByName(engineName, false)
			if engine == nil {
				return fmt.Errorf("unknown search engine %q. Available: %s",
					engineName, strings.Join(search.AvailableEngines(), ", "))
			}

			results, err := engine.Search(term)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}
			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			// Display results
			fmt.Printf("\nFound %d result(s):\n\n", len(results))
			for i, pack := range results {
				fmt.Printf("  [%3d] %s [%s]\n", i+1,
					pack.Filename,
					entities.HumanReadableBytes(pack.Size))
			}

			// Interactive selection
			selected, err := selectPacks(results)
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				fmt.Println("No packs selected.")
				return nil
			}

			entities.PreparePacks(selected, out)

			throttleBytes, err := entities.ParseThrottle(throttle)
			if err != nil {
				return fmt.Errorf("invalid throttle value %q: %w", throttle, err)
			}

			downloader.DownloadPacks(selected, downloader.Options{
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

	cmd.Flags().StringVar(&engineName, "search-engine", "xdcc-eu",
		"Search engine to use (nibl, xdcc-eu, ixirc, subsplease)")
	cmd.Flags().StringVarP(&server, "server", "s", "irc.rizon.net",
		"IRC server (overridden by search result if different)")
	cmd.Flags().StringVarP(&out, "out", "o", "",
		"Output file path (defaults to pack filename)")
	cmd.Flags().StringVarP(&throttle, "throttle", "t", "-1",
		"Download speed limit (e.g. 50M, 100K). -1 = unlimited")
	cmd.Flags().IntVar(&connectTimeout, "connect-timeout", 120,
		"Seconds to wait for the bot to initiate DCC (before transfer starts)")
	cmd.Flags().IntVar(&stallTimeout, "stall-timeout", 60,
		"Seconds of no transfer progress before aborting (0 = disabled)")
	cmd.Flags().StringVar(&fallbackChannel, "fallback-channel", "",
		"IRC channel to join if WHOIS finds none")
	cmd.Flags().IntVar(&waitTime, "wait-time", 0,
		"Seconds to wait before sending the XDCC request")
	cmd.Flags().StringVar(&username, "username", "",
		"IRC nickname to use (random if not set)")
	cmd.Flags().IntVar(&channelJoinDelay, "channel-join-delay", -1,
		"Seconds to wait after connecting (-1 = random 5-10s)")
	cmd.Flags().CountVarP(&verbosity, "verbose", "v", "Verbose output (-v=verbose, -vv=debug)")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode")
	_ = server

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// verbosityLevel converts count+quiet flags to verbosity int.
func verbosityLevel(count int, quiet bool) int {
	if quiet {
		return -1
	}
	return count
}

// selectPacks prompts the user to select one or more packs from the results list.
// Accepts: single number (3), range (1-5), comma list (1,3,5), or "all".
func selectPacks(results []*entities.XDCCPack) ([]*entities.XDCCPack, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter selection (number, range 1-5, list 1,3,5, or 'all'): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	input = strings.TrimSpace(input)

	if strings.ToLower(input) == "all" {
		return results, nil
	}

	var selected []*entities.XDCCPack
	seen := make(map[int]bool)

	addIdx := func(i int) {
		if i >= 1 && i <= len(results) && !seen[i] {
			seen[i] = true
			selected = append(selected, results[i-1])
		}
	}

	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, e1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, e2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if e1 != nil || e2 != nil {
				return nil, fmt.Errorf("invalid selection: %s", part)
			}
			for i := start; i <= end; i++ {
				addIdx(i)
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid selection: %s", part)
			}
			addIdx(n)
		}
	}

	return selected, nil
}
