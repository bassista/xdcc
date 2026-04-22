package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
		extFilter        string
		botFilter        string
	)

	cmd := &cobra.Command{
		Use:   "xdcc-browse <search_term>",
		Short: "Search for XDCC packs and download interactively",
		Long: `xdcc-browse searches for XDCC packs, optionally filters the results,
displays a numbered list, and then downloads the selected pack(s).

Filters (applied before the selection menu):
  --ext   keep only files with the given extension(s)  (e.g. --ext=mkv,avi)
  --bot   keep only packs from bots whose name contains the given substring

Selection syntax (after the list is shown):
  3        single pack
  1-5      range (packs 1 through 5)
  1,3,7    comma-separated list
  all      download everything in the list

Verbosity levels:
  (default)  show connection and download progress
  -v         also show bot notices, channel joins, WHOIS results
  -vv        full debug (DNS, DCC internals, all IRC events)
  -q         suppress all output`,
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

			// Filter by extension if requested
			if extFilter != "" {
				results = filterByExtension(results, extFilter)
			}

			// Filter by bot name if requested
			if botFilter != "" {
				results = filterByBot(results, botFilter)
			}

			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			// Display results
			fmt.Printf("\nFound %d result(s):\n\n", len(results))
			for i, pack := range results {
				fmt.Printf("  [%3d] %s [%s] bot: %s\n", i+1,
					pack.Filename,
					entities.HumanReadableBytes(pack.Size),
					pack.Bot)
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

			// If --server was explicitly set, override the server on all selected packs
			if server != "" {
				srv := entities.ParseIrcServer(server)
				for _, p := range selected {
					p.Server = srv
				}
			}

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
		"Search engine to use: nibl, xdcc-eu, ixirc, subsplease (default: xdcc-eu)")
	cmd.Flags().StringVarP(&server, "server", "s", "",
		"Override IRC server for all selected packs (host or host:port). Default: use server from search result")
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
	cmd.Flags().StringVar(&extFilter, "ext", "",
		"Filter results by file extension(s), comma-separated (e.g. --ext=mkv,avi,mp4)")
	cmd.Flags().StringVar(&botFilter, "bot", "",
		"Filter results by bot name substring, case-insensitive (e.g. --bot=WOND)")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// filterByBot returns only packs whose bot name contains the given substring (case-insensitive).
func filterByBot(packs []*entities.XDCCPack, substr string) []*entities.XDCCPack {
	sub := strings.ToLower(substr)
	var out []*entities.XDCCPack
	for _, p := range packs {
		if strings.Contains(strings.ToLower(p.Bot), sub) {
			out = append(out, p)
		}
	}
	return out
}
// extList is a comma-separated string like "mkv,avi,mp4".
func filterByExtension(packs []*entities.XDCCPack, extList string) []*entities.XDCCPack {
	exts := make(map[string]bool)
	for _, e := range strings.Split(extList, ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" {
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			exts[e] = true
		}
	}
	var out []*entities.XDCCPack
	for _, p := range packs {
		ext := strings.ToLower(filepath.Ext(p.Filename))
		if exts[ext] {
			out = append(out, p)
		}
	}
	return out
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
