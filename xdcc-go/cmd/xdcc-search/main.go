package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"xdcc-go/internal/entities"
	"xdcc-go/internal/search"
)

func main() {
	var engineName string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "xdcc-search <search_term> <engine>",
		Short: "Search for XDCC packs",
		Long: `xdcc-search searches for XDCC packs using the specified search engine
and prints the results with the corresponding download commands.

Available engines: ` + strings.Join(search.AvailableEngines(), ", "),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			term := args[0]
			engineName = args[1]

			engine := search.EngineByName(engineName, verbose)
			if engine == nil {
				return fmt.Errorf("unknown search engine %q. Available: %s",
					engineName, strings.Join(search.AvailableEngines(), ", "))
			}

			results, err := engine.Search(term)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(results) == 0 {
				fmt.Fprintln(os.Stderr, "No results found.")
				return nil
			}

			for _, pack := range results {
				msg := pack.GetRequestMessage(true)
				line := fmt.Sprintf("%s [%s] (xdcc-dl \"%s\")",
					pack.Filename,
					entities.HumanReadableBytes(pack.Size),
					msg)
				if pack.Server.Address != "irc.rizon.net" {
					line = fmt.Sprintf("%s [%s] (xdcc-dl \"%s\" --server %s)",
						pack.Filename,
						entities.HumanReadableBytes(pack.Size),
						msg,
						pack.Server.Address)
				}
				fmt.Println(line)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose/debug output")

	_ = engineName
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
