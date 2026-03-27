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
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read instance URL: %w", err)
				}
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
			authCode, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read authorization code: %w", err)
			}
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
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read status: %w", err)
				}
				status = strings.TrimSpace(input)
			}

			if status == "" {
				return fmt.Errorf("status cannot be empty")
			}

			s, err := client.PostStatus(status, visibility)
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

func GetTimelineCommand() *cobra.Command {
	var timeline string
	var limit int

	cmd := &cobra.Command{
		Use:   "timeline [home|local|federated]",
		Short: "View timeline",
		Long:  `View timeline posts. Use 'home' for home timeline, 'local' for local timeline, or 'federated' for federated timeline.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			if len(args) > 0 {
				timeline = args[0]
			}

			if timeline == "" {
				timeline = "home"
			}

			var statuses []api.Status
			var err error

			switch timeline {
			case "home":
				statuses, err = client.GetHomeTimeline(limit)
			case "local":
				statuses, err = client.GetLocalTimeline(limit)
			case "federated", "public":
				statuses, err = client.GetFederatedTimeline(limit)
			default:
				return fmt.Errorf("invalid timeline type: %s (use: home, local, or federated)", timeline)
			}

			if err != nil {
				return fmt.Errorf("failed to get timeline: %w", err)
			}

			if len(statuses) == 0 {
				fmt.Println("No posts found.")
				return nil
			}

			fmt.Printf("=== %s timeline ===\n\n", timeline)
			for i, s := range statuses {
				fmt.Printf("[%d] @%s\n", i+1, s.Account.Username)
				fmt.Printf("    %s\n", stripHTML(s.Content))
				fmt.Printf("    %s\n", s.URL)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of posts to fetch")

	return cmd
}

func GetStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <status_id>",
		Short: "View a status",
		Long:  "View details of a specific status by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			statusID := args[0]

			status, err := client.GetStatus(statusID)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			fmt.Printf("Author: @%s\n", status.Account.Username)
			if status.Account.DisplayName != "" {
				fmt.Printf("Display Name: %s\n", status.Account.DisplayName)
			}
			fmt.Printf("Created: %s\n", status.CreatedAt)
			fmt.Printf("URL: %s\n", status.URL)
			fmt.Printf("\nContent:\n%s\n", stripHTML(status.Content))

			return nil
		},
	}

	return cmd
}

func GetFavouriteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "favourite <status_id>",
		Short: "Favourite a status",
		Long:  "Favourite (like) a status by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			statusID := args[0]

			_, err := client.FavoriteStatus(statusID)
			if err != nil {
				return fmt.Errorf("failed to favourite status: %w", err)
			}

			fmt.Println("Status favourited!")
			return nil
		},
	}

	return cmd
}

func GetUnfavouriteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfavourite <status_id>",
		Short: "Unfavourite a status",
		Long:  "Remove favourite from a status by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			statusID := args[0]

			_, err := client.UnfavoriteStatus(statusID)
			if err != nil {
				return fmt.Errorf("failed to unfavourite status: %w", err)
			}

			fmt.Println("Status unfavourited!")
			return nil
		},
	}

	return cmd
}

func GetBoostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "boost <status_id>",
		Short: "Boost (reblog) a status",
		Long:  "Boost (reblog) a status by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			statusID := args[0]

			_, err := client.BoostStatus(statusID)
			if err != nil {
				return fmt.Errorf("failed to boost status: %w", err)
			}

			fmt.Println("Status boosted!")
			return nil
		},
	}

	return cmd
}

func GetUnboostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unboost <status_id>",
		Short: "Unboost (unreblog) a status",
		Long:  "Remove boost from a status by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			statusID := args[0]

			_, err := client.UnboostStatus(statusID)
			if err != nil {
				return fmt.Errorf("failed to unboost status: %w", err)
			}

			fmt.Println("Status unboosted!")
			return nil
		},
	}

	return cmd
}

func GetReplyCommand() *cobra.Command {
	var inReplyToID string
	var visibility string

	cmd := &cobra.Command{
		Use:   "reply <status_id> <message>",
		Short: "Reply to a status",
		Long:  "Reply to a status by providing the status ID and your reply message.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			inReplyToID = args[0]
			status := strings.Join(args[1:], " ")

			s, err := client.PostReply(status, inReplyToID, visibility)
			if err != nil {
				return fmt.Errorf("failed to post reply: %w", err)
			}

			fmt.Printf("Reply posted!\n")
			fmt.Printf("URL: %s\n", s.URL)
			return nil
		},
	}

	cmd.Flags().StringVarP(&visibility, "visibility", "v", "public", "Reply visibility (public, unlisted, private, direct)")

	return cmd
}

func GetDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <status_id>",
		Short: "Delete your own status",
		Long:  "Delete one of your own statuses by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			statusID := args[0]

			err := client.DeleteStatus(statusID)
			if err != nil {
				return fmt.Errorf("failed to delete status: %w", err)
			}

			fmt.Println("Status deleted!")
			return nil
		},
	}

	return cmd
}

func GetSearchCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for users or statuses",
		Long:  "Search for users, statuses, and hashtags on Mastodon.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			query := args[0]

			result, err := client.Search(query, limit)
			if err != nil {
				return fmt.Errorf("failed to search: %w", err)
			}

			if len(result.Accounts) == 0 && len(result.Statuses) == 0 && len(result.Hashtags) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			if len(result.Accounts) > 0 {
				fmt.Println("=== Accounts ===")
				for _, account := range result.Accounts {
					fmt.Printf("@%s", account.Username)
					if account.DisplayName != "" {
						fmt.Printf(" (%s)", account.DisplayName)
					}
					fmt.Println()
				}
				fmt.Println()
			}

			if len(result.Statuses) > 0 {
				fmt.Println("=== Statuses ===")
				for _, s := range result.Statuses {
					fmt.Printf("@%s: %s\n", s.Account.Username, stripHTML(s.Content))
					fmt.Printf("    %s\n", s.URL)
				}
				fmt.Println()
			}

			if len(result.Hashtags) > 0 {
				fmt.Println("=== Hashtags ===")
				for _, tag := range result.Hashtags {
					fmt.Printf("#%s\n", tag)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Number of results to fetch")

	return cmd
}

func GetAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account <username|account_id>",
		Short: "View user account info",
		Long:  "View detailed information about a user by username or account ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			usernameOrID := args[0]

			var account *api.AccountFull
			var err error

			if isID(usernameOrID) {
				account, err = client.GetAccount(usernameOrID)
				if err != nil {
					return fmt.Errorf("failed to get account: %w", err)
				}
			} else {
				var acc *api.Account
				acc, err = client.GetAccountByUsername(usernameOrID)
				if err != nil {
					return fmt.Errorf("failed to find user: %w", err)
				}
				account, err = client.GetAccount(acc.ID)
				if err != nil {
					return fmt.Errorf("failed to get account info: %w", err)
				}
			}

			fmt.Printf("Username: @%s\n", account.Username)
			if account.DisplayName != "" {
				fmt.Printf("Display Name: %s\n", account.DisplayName)
			}
			fmt.Printf("Account: %s\n", account.Acct)
			fmt.Printf("URL: %s\n", account.URL)
			fmt.Printf("Created: %s\n", account.CreatedAt)
			fmt.Printf("Followers: %d\n", account.FollowersCount)
			fmt.Printf("Following: %d\n", account.FollowingCount)
			fmt.Printf("Statuses: %d\n", account.StatusesCount)
			if account.Locked {
				fmt.Println("Locked: yes")
			}
			if account.Bot {
				fmt.Println("Bot: yes")
			}
			if account.Note != "" {
				fmt.Printf("\nNote:\n%s\n", stripHTML(account.Note))
			}
			if len(account.Fields) > 0 {
				fmt.Println("\nFields:")
				for _, field := range account.Fields {
					fmt.Printf("  %s: %s\n", field.Name, stripHTML(field.Value))
				}
			}

			return nil
		},
	}

	return cmd
}

func GetNotificationsCommand() *cobra.Command {
	var limit int
	var notificationType string

	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "View notifications",
		Long:  "View your notifications (mentions, favourites, boosts, follows).",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !api.IsLoggedIn() {
				return fmt.Errorf("please login first using: mastodon-cli login")
			}

			cfg := api.GetConfig()
			client := api.NewClient()
			client.SetInstanceURL(cfg.InstanceURL)
			client.SetAccessToken(cfg.AccessToken)

			notifications, err := client.GetNotifications(limit, notificationType)
			if err != nil {
				return fmt.Errorf("failed to get notifications: %w", err)
			}

			if len(notifications) == 0 {
				fmt.Println("No notifications.")
				return nil
			}

			fmt.Println("=== Notifications ===\n")
			for _, n := range notifications {
				switch n.Type {
				case "mention":
					fmt.Printf("[%s] @%s mentioned you\n", n.CreatedAt, n.Account.Username)
					if n.Status != nil {
						fmt.Printf("    %s\n", stripHTML(n.Status.Content))
					}
				case "favourite":
					fmt.Printf("[%s] @%s favourited your post\n", n.CreatedAt, n.Account.Username)
				case "reblog":
					fmt.Printf("[%s] @%s boosted your post\n", n.CreatedAt, n.Account.Username)
				case "follow":
					fmt.Printf("[%s] @%s followed you\n", n.CreatedAt, n.Account.Username)
				case "follow_request":
					fmt.Printf("[%s] @%s wants to follow you\n", n.CreatedAt, n.Account.Username)
				default:
					fmt.Printf("[%s] @%s: %s\n", n.CreatedAt, n.Account.Username, n.Type)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of notifications to fetch")
	cmd.Flags().StringVarP(&notificationType, "type", "t", "", "Filter by type (mention, favourite, reblog, follow)")

	return cmd
}

func GetFollowersCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "followers <username>",
		Short: "View followers",
		Long:  "View the list of followers for a user.",
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

			followers, err := client.GetAccountFollowers(account.ID, limit)
			if err != nil {
				return fmt.Errorf("failed to get followers: %w", err)
			}

			if len(followers) == 0 {
				fmt.Println("No followers found.")
				return nil
			}

			fmt.Printf("=== Followers of @%s ===\n\n", username)
			for _, f := range followers {
				fmt.Printf("@%s", f.Username)
				if f.DisplayName != "" {
					fmt.Printf(" (%s)", f.DisplayName)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of followers to fetch")

	return cmd
}

func GetFollowingCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "following <username>",
		Short: "View following",
		Long:  "View the list of users that a user is following.",
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

			following, err := client.GetAccountFollowing(account.ID, limit)
			if err != nil {
				return fmt.Errorf("failed to get following: %w", err)
			}

			if len(following) == 0 {
				fmt.Println("Not following anyone.")
				return nil
			}

			fmt.Printf("=== Users @%s is following ===\n\n", username)
			for _, f := range following {
				fmt.Printf("@%s", f.Username)
				if f.DisplayName != "" {
					fmt.Printf(" (%s)", f.DisplayName)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of following to fetch")

	return cmd
}

func stripHTML(s string) string {
	re := strings.NewReplacer("<p>", "\n", "</p>", "", "<br>", "\n", "<br />", "\n", "&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", "\"", "&#39;", "'")
	return re.Replace(s)
}

func isID(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}
