/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"deadlinks/internal/crawler"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deadlinks --url <URL>",
	Short: "Search for dead links within a domain",
	Long: `Dead Links is a CLI tool that crawls a specified domain to discover
and report broken or dead links. It recursively checks all links found
within the domain and identifies those that return error responses.`,

	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		// fmt.Println("URL:", url)
		if url == "" {
			fmt.Fprintf(os.Stderr, "Error: url flag is required\n")
			cmd.Help()
			os.Exit(1)
		}

		crawler.Init(url)
		deadLinks := crawler.Start()

		if len(deadLinks) == 0 {
			fmt.Println("\033[32m\n✅ No dead links found!")
			return
		}

		totalDeadLinks := 0
		for _, links := range deadLinks {
			totalDeadLinks += len(links)
		}

		fmt.Printf("\n\033[31mFound %d dead links:\n\n\033[97m", totalDeadLinks)

		w := tabwriter.NewWriter(os.Stdout, 10, 0, 3, ' ', 0)
		defer w.Flush()

		fmt.Fprintln(w, "\033[97mPAGE\tLINK\tSTATUS\033[0m")
		fmt.Fprintln(w, "\033[97m----\t---------\t------\033[0m")

		for page, links := range deadLinks {
			for _, link := range links {
				if link.Internal {
					fmt.Fprintf(w, "\033[31m%s\t%s\tDead\033[0m\n", page, link.Link)
				} else {
					fmt.Fprintf(w, "\033[33m%s\t%s\tMaybe dead\033[0m\n", page, link.Link)
				}
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("url", "u", "", "URL of the site to scrape (required)")
	rootCmd.MarkFlagRequired("url")
}
