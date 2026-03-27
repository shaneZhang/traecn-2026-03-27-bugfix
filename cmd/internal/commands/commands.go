package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mastodon-cli/cmd/internal/api"

	"github.com/spf13/cobra"
)

func GetLoginCommand() *cobra.Command {
	var instance string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "login [instance]",
		Short: "Login to a Mastodon instance",
		Long: `Login to a Mastodon instance using OAuth2 authentication.
If no instance is provided, it will be prompted interactively.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				instance = args[0]
			}

			if instance == "" && !interactive {
				fmt.Print("Enter Mastodon instance URL (e.g., mastodon.social): ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				instance = strings.TrimSpace(input)
			}

			if instance == "" {
				return fmt.Errorf("instance URL is required")
			}

			fmt.Printf("Registering application on %s...\n", instance)

			app, err := api.RegisterApp(instance)
			if err != nil {
				return fmt.Errorf("failed to register app: %w", err)
			}

			fmt.Println("Application registered successfully!")

			authURL, err := api.GetAuthorizationURL(instance, app.ClientID, app.ClientSecret)
			if err != nil {
				return fmt.Errorf("failed to get authorization URL: %w", err)
			}

			fmt.Println("\nPlease open the following URL in your browser:")
			fmt.Println(authURL)
			fmt.Println("\nOr the CLI will attempt to open it automatically...")

			if err := api.OpenURL(authURL); err != nil {
				fmt.Println("Could not open browser automatically. Please copy the URL above.")
			}

			fmt.Print("\nEnter the authorization code: ")
			reader := bufio.NewReader(os.Stdin)
			authCode, _ := reader.ReadString('\n')
			authCode = strings.TrimSpace(authCode)

			if authCode == "" {
				return fmt.Errorf("authorization code is required")
			}

			accessToken, err := api.Login(instance, app.ClientID, app.ClientSecret, authCode)
			if err != nil {
				return fmt.Errorf("failed to login: %w", err)
			}

			if err := api.SaveLogin(instance, accessToken, app.ClientID, app.ClientSecret); err != nil {
				return fmt.Errorf("failed to save login: %w", err)
			}

			fmt.Println("\nLogin successful!")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "I", false, "Interactive mode")
	cmd.Flags().StringVarP(&instance, "instance", "", "", "Mastodon instance URL")

	return cmd
}

func GetLogoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Mastodon",
		Long:  "Clear stored credentials and logout from Mastodon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				fmt.Println("You are not logged in.")
				return nil
			}

			if err := api.Logout(); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}

			fmt.Println("Logged out successfully!")
			return nil
		},
	}

	return cmd
}

func GetPostCommand() *cobra.Command {
	var visibility string

	cmd := &cobra.Command{
		Use:   "post [status]",
		Short: "Post a status to Mastodon",
		Long:  `Post a new status (toot) to Mastodon. If no status is provided, it will be read from stdin.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			var status string
			if len(args) > 0 {
				status = strings.Join(args, " ")
			} else {
				fmt.Print("Enter your status: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				status = strings.TrimSpace(input)
			}

			if status == "" {
				return fmt.Errorf("status cannot be empty")
			}

			s, err := client.PostStatus(status)
			if err != nil {
				return fmt.Errorf("failed to post: %w", err)
			}

			fmt.Printf("Posted successfully!\n")
			fmt.Printf("URL: %s\n", s.URL)
			return nil
		},
	}

	cmd.Flags().StringVarP(&visibility, "visibility", "v", "public", "Post visibility (public, unlisted, private, direct)")

	return cmd
}

func GetFollowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "follow <username>",
		Short: "Follow a user",
		Long:  "Follow a user on Mastodon. Use the username without @ or the full account address.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			username := args[0]

			account, err := client.GetAccountByUsername(username)
			if err != nil {
				return fmt.Errorf("failed to find user: %w", err)
			}

			rel, err := client.FollowAccount(account.ID)
			if err != nil {
				return fmt.Errorf("failed to follow user: %w", err)
			}

			if rel.Following {
				fmt.Printf("You are now following @%s\n", username)
			} else {
				fmt.Printf("Follow request sent to @%s\n", username)
			}

			return nil
		},
	}

	return cmd
}

func GetUnfollowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfollow <username>",
		Short: "Unfollow a user",
		Long:  "Unfollow a user on Mastodon. Use the username without @ or the full account address.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			username := args[0]

			account, err := client.GetAccountByUsername(username)
			if err != nil {
				return fmt.Errorf("failed to find user: %w", err)
			}

			_, err = client.UnfollowAccount(account.ID)
			if err != nil {
				return fmt.Errorf("failed to unfollow user: %w", err)
			}

			fmt.Printf("You have unfollowed @%s\n", username)

			return nil
		},
	}

	return cmd
}

func GetWhoamiCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current logged in user",
		Long:  "Display information about the currently logged in user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			account, err := client.VerifyCredentials()
			if err != nil {
				return fmt.Errorf("failed to get account info: %w", err)
			}

			fmt.Printf("Logged in as: @%s\n", account.Acct)
			if account.DisplayName != "" {
				fmt.Printf("Display name: %s\n", account.DisplayName)
			}
			fmt.Printf("Account ID: %s\n", account.ID)

			return nil
		},
	}

	return cmd
}
