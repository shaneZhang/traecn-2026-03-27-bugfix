package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"mastodon-cli/cmd/internal/config"
)

type Client struct {
	httpClient  *http.Client
	instanceURL string
	accessToken string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) SetInstanceURL(instanceURL string) {
	c.instanceURL = instanceURL
}

func (c *Client) SetAccessToken(accessToken string) {
	c.accessToken = accessToken
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	baseURL := c.instanceURL
	if !hasScheme(baseURL) {
		baseURL = "https://" + baseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid instance URL: %w", err)
	}

	parsedURL.Path = "/api/v1/" + endpoint
	apiURL := parsedURL.String()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, apiURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func hasScheme(url string) bool {
	return len(url) >= 7 && (url[:7] == "http://" || url[:8] == "https://")
}

type AppRegistration struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func RegisterApp(instanceURL string) (*AppRegistration, error) {
	baseURL := instanceURL
	if !hasScheme(baseURL) {
		baseURL = "https://" + baseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid instance URL: %w", err)
	}
	parsedURL.Path = "/api/v1/apps"

	data := url.Values{}
	data.Set("client_name", "mastodon-cli")
	data.Set("redirect_uris", "urn:ietf:wg:oauth:2.0:oob")
	data.Set("scopes", "read write follow push")

	resp, err := http.PostForm(parsedURL.String(), data)
	if err != nil {
		return nil, fmt.Errorf("failed to register app: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to register app: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &AppRegistration{
		ClientID:     result.ClientID,
		ClientSecret: result.ClientSecret,
	}, nil
}

func GetAuthorizationURL(instanceURL, clientID, clientSecret string) (string, error) {
	baseURL := instanceURL
	if !hasScheme(baseURL) {
		baseURL = "https://" + baseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid instance URL: %w", err)
	}
	parsedURL.Path = "/oauth/authorize"

	query := parsedURL.Query()
	query.Set("client_id", clientID)
	query.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	query.Set("response_type", "code")
	query.Set("scope", "read write follow push")
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func GetAccessToken(instanceURL, clientID, clientSecret, authCode string) (string, error) {
	baseURL := instanceURL
	if !hasScheme(baseURL) {
		baseURL = "https://" + baseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid instance URL: %w", err)
	}
	parsedURL.Path = "/oauth/token"

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)
	data.Set("scope", "read write follow push")

	resp, err := http.PostForm(parsedURL.String(), data)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get access token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.AccessToken, nil
}

func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

type Status struct {
	ID        string  `json:"id"`
	Content   string  `json:"content"`
	URL       string  `json:"url"`
	CreatedAt string  `json:"created_at"`
	Account   Account `json:"account"`
}

type Account struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Acct        string `json:"acct"`
	DisplayName string `json:"display_name"`
}

type Relationship struct {
	ID         string `json:"id"`
	Following  bool   `json:"following"`
	FollowedBy bool   `json:"followed_by"`
}

func (c *Client) PostStatus(status string) (*Status, error) {
	body := map[string]interface{}{
		"status": status,
	}

	respBody, err := c.doRequest("POST", "statuses", body)
	if err != nil {
		return nil, err
	}

	var s Status
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &s, nil
}

func (c *Client) GetAccountByUsername(username string) (*Account, error) {
	respBody, err := c.doRequest("GET", "accounts/lookup?acct="+url.QueryEscape(username), nil)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := json.Unmarshal(respBody, &account); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &account, nil
}

func (c *Client) FollowAccount(accountID string) (*Relationship, error) {
	respBody, err := c.doRequest("POST", "accounts/"+accountID+"/follow", nil)
	if err != nil {
		return nil, err
	}

	var rel Relationship
	if err := json.Unmarshal(respBody, &rel); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &rel, nil
}

func (c *Client) UnfollowAccount(accountID string) (*Relationship, error) {
	respBody, err := c.doRequest("POST", "accounts/"+accountID+"/unfollow", nil)
	if err != nil {
		return nil, err
	}

	var rel Relationship
	if err := json.Unmarshal(respBody, &rel); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &rel, nil
}

func (c *Client) VerifyCredentials() (*Account, error) {
	respBody, err := c.doRequest("GET", "accounts/verify_credentials", nil)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := json.Unmarshal(respBody, &account); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &account, nil
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	CreatedAt   int    `json:"created_at"`
}

func Login(instanceURL, clientID, clientSecret, authCode string) (string, error) {
	baseURL := instanceURL
	if !hasScheme(baseURL) {
		baseURL = "https://" + baseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid instance URL: %w", err)
	}
	parsedURL.Path = "/oauth/token"

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)
	data.Set("scope", "read write follow push")

	resp, err := http.PostForm(parsedURL.String(), data)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get access token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.AccessToken, nil
}

func SaveLogin(instanceURL, accessToken, clientID, clientSecret string) error {
	cfg := &config.Config{
		InstanceURL:  instanceURL,
		AccessToken:  accessToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return config.SaveConfig(cfg)
}

func IsLoggedIn() bool {
	return config.IsLoggedIn()
}

func Logout() error {
	return config.ClearConfig()
}

func GetConfig() *config.Config {
	return config.GetConfig()
}
