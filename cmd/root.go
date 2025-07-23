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
		fmt.Println("URL:", url)
		if url == "" {
			fmt.Fprintf(os.Stderr, "Error: url flag is required\n")
			cmd.Help()
			os.Exit(1)
		}

		crawler.Init(url)
		deadLinks := crawler.Start()

		if len(deadLinks) == 0 {
			fmt.Println("✅ No dead links found!")
			return
		}

		totalDeadLinks := 0
		for _, links := range deadLinks {
			totalDeadLinks += len(links)
		}

		fmt.Printf("\n\033[31mFound %d dead links:\n\n\033[97m", totalDeadLinks)

		w := tabwriter.NewWriter(os.Stdout, 10, 0, 3, ' ', 0)
		defer w.Flush()

		fmt.Fprintln(w, "PAGE\tLINK\tSTATUS")
		fmt.Fprintln(w, "----\t---------\t------")

		for page, links := range deadLinks {
			for _, link := range links {
				fmt.Fprintf(w, "%s\t%s\t❌ Dead\n", page, link)
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
