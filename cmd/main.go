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
	rootCmd.AddCommand(commands.GetTimelineCommand())
	rootCmd.AddCommand(commands.GetStatusCommand())
	rootCmd.AddCommand(commands.GetFavouriteCommand())
	rootCmd.AddCommand(commands.GetUnfavouriteCommand())
	rootCmd.AddCommand(commands.GetBoostCommand())
	rootCmd.AddCommand(commands.GetUnboostCommand())
	rootCmd.AddCommand(commands.GetReplyCommand())
	rootCmd.AddCommand(commands.GetDeleteCommand())
	rootCmd.AddCommand(commands.GetSearchCommand())
	rootCmd.AddCommand(commands.GetAccountCommand())
	rootCmd.AddCommand(commands.GetNotificationsCommand())
	rootCmd.AddCommand(commands.GetFollowersCommand())
	rootCmd.AddCommand(commands.GetFollowingCommand())

	// Check login status before executing command
	cfg := api.GetConfig()
	if cfg.InstanceURL != "" && cfg.AccessToken == "" {
		fmt.Fprintln(os.Stderr, "Warning: Logged in but no access token. Please run login again.")
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
