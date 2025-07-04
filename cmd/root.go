/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

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

		fmt.Printf("Scanning %s for dead links...\n", url)
		//logic
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
