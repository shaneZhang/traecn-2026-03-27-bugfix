package main

import (
	"fmt"
	"os"

	"mastodon-cli/cmd/internal/api"
	"mastodon-cli/cmd/internal/commands"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mastodon-cli",
		Short: "Mastodon CLI - A command-line interface for Mastodon",
		Long:  `A command-line tool to interact with Mastodon social network. Supports posting, following, and more.`,
	}

	rootCmd.AddCommand(commands.GetLoginCommand())
	rootCmd.AddCommand(commands.GetLogoutCommand())
	rootCmd.AddCommand(commands.GetPostCommand())
	rootCmd.AddCommand(commands.GetFollowCommand())
	rootCmd.AddCommand(commands.GetUnfollowCommand())
	rootCmd.AddCommand(commands.GetWhoamiCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg := api.GetConfig()
	if cfg.InstanceURL != "" && cfg.AccessToken == "" {
		fmt.Println("Warning: Logged in but no access token. Please run login again.")
	}
}
