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
	"strings"
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

	// 确保 baseURL 没有尾部斜杠
	baseURL = strings.TrimSuffix(baseURL, "/")

	var apiURL string
	// 如果 endpoint 以 /api/ 开头，则使用完整路径（用于跨版本API调用）
	if strings.HasPrefix(endpoint, "/api/") {
		apiURL = baseURL + endpoint
	} else {
		// 默认使用 /api/v1/ 前缀
		apiURL = baseURL + "/api/v1/" + endpoint
	}

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
	req.Header.Set("Accept", "application/json")
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
		// 如果返回的是HTML，提供更友好的错误信息
		if strings.Contains(string(respBody), "<!DOCTYPE html>") || strings.Contains(string(respBody), "<html") {
			return nil, fmt.Errorf("API request failed with status %d: Server returned HTML instead of JSON. This may indicate:\n"+
				"  - The instance URL is incorrect (currently using: %s)\n"+
				"  - The API endpoint is not supported on this instance\n"+
				"  - The instance may require authentication for this endpoint",
				resp.StatusCode, baseURL)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func hasScheme(url string) bool {
	if len(url) < 7 {
		return false
	}
	if url[:7] == "http://" {
		return true
	}
	if len(url) >= 8 && url[:8] == "https://" {
		return true
	}
	return false
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

func (c *Client) PostStatus(status string, visibility string) (*Status, error) {
	body := map[string]interface{}{
		"status": status,
	}

	if visibility != "" {
		body["visibility"] = visibility
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
	// 首先尝试使用 accounts/lookup 端点
	endpoint := "accounts/lookup?acct=" + url.QueryEscape(username)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		// 如果 lookup 失败，尝试搜索端点作为备选
		searchEndpoint := "accounts/search?q=" + url.QueryEscape(username) + "&limit=1&resolve=true"
		searchRespBody, searchErr := c.doRequest("GET", searchEndpoint, nil)
		if searchErr != nil {
			return nil, fmt.Errorf("failed to find user (lookup: %w, search: %w)", err, searchErr)
		}

		var accounts []Account
		if jsonErr := json.Unmarshal(searchRespBody, &accounts); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse search response: %w", jsonErr)
		}

		if len(accounts) == 0 {
			return nil, fmt.Errorf("no user found with username: %s", username)
		}

		return &accounts[0], nil
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

// Login 已弃用，请使用 GetAccessToken 代替
func Login(instanceURL, clientID, clientSecret, authCode string) (string, error) {
	return GetAccessToken(instanceURL, clientID, clientSecret, authCode)
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

func (c *Client) GetHomeTimeline(limit int) ([]Status, error) {
	endpoint := "timelines/home"
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var statuses []Status
	if err := json.Unmarshal(respBody, &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return statuses, nil
}

func (c *Client) GetLocalTimeline(limit int) ([]Status, error) {
	endpoint := "timelines/public"
	query := url.Values{}
	query.Set("local", "true")
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	endpoint += "?" + query.Encode()
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var statuses []Status
	if err := json.Unmarshal(respBody, &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return statuses, nil
}

func (c *Client) GetFederatedTimeline(limit int) ([]Status, error) {
	endpoint := "timelines/public"
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var statuses []Status
	if err := json.Unmarshal(respBody, &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return statuses, nil
}

func (c *Client) GetStatus(statusID string) (*Status, error) {
	respBody, err := c.doRequest("GET", "statuses/"+statusID, nil)
	if err != nil {
		return nil, err
	}

	var s Status
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &s, nil
}

func (c *Client) FavoriteStatus(statusID string) (*Status, error) {
	respBody, err := c.doRequest("POST", "statuses/"+statusID+"/favourite", nil)
	if err != nil {
		return nil, err
	}

	var s Status
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &s, nil
}

func (c *Client) UnfavoriteStatus(statusID string) (*Status, error) {
	respBody, err := c.doRequest("POST", "statuses/"+statusID+"/unfavourite", nil)
	if err != nil {
		return nil, err
	}

	var s Status
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &s, nil
}

// FavouriteStatus 是 FavoriteStatus 的别名，保持英式拼写的兼容性
func (c *Client) FavouriteStatus(statusID string) (*Status, error) {
	return c.FavoriteStatus(statusID)
}

// UnfavouriteStatus 是 UnfavoriteStatus 的别名，保持英式拼写的兼容性
func (c *Client) UnfavouriteStatus(statusID string) (*Status, error) {
	return c.UnfavoriteStatus(statusID)
}

func (c *Client) BoostStatus(statusID string) (*Status, error) {
	respBody, err := c.doRequest("POST", "statuses/"+statusID+"/reblog", nil)
	if err != nil {
		return nil, err
	}

	var s Status
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &s, nil
}

func (c *Client) UnboostStatus(statusID string) (*Status, error) {
	respBody, err := c.doRequest("POST", "statuses/"+statusID+"/unreblog", nil)
	if err != nil {
		return nil, err
	}

	var s Status
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &s, nil
}

func (c *Client) PostReply(status, inReplyToID, visibility string) (*Status, error) {
	body := map[string]interface{}{
		"status":      status,
		"in_reply_to": inReplyToID,
	}

	if visibility != "" {
		body["visibility"] = visibility
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

func (c *Client) DeleteStatus(statusID string) error {
	_, err := c.doRequest("DELETE", "statuses/"+statusID, nil)
	if err != nil {
		return err
	}

	return nil
}

type Tag struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SearchResult struct {
	Accounts []Account `json:"accounts"`
	Statuses []Status  `json:"statuses"`
	Hashtags []Tag     `json:"hashtags"`
}

func (c *Client) Search(query string, limit int) (*SearchResult, error) {
	queryParams := url.Values{}
	queryParams.Set("q", query)
	queryParams.Set("resolve", "true")
	if limit > 0 {
		queryParams.Set("limit", fmt.Sprintf("%d", limit))
	}
	queryString := "?" + queryParams.Encode()

	// 首先尝试 v2 搜索端点
	endpoint := "/api/v2/search" + queryString
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		// 回退到 v1 搜索端点
		endpoint = "search" + queryString
		respBody, err = c.doRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
	}

	var result SearchResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

type AccountFull struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Acct           string  `json:"acct"`
	DisplayName    string  `json:"display_name"`
	Note           string  `json:"note"`
	URL            string  `json:"url"`
	Avatar         string  `json:"avatar"`
	AvatarStatic   string  `json:"avatar_static"`
	Header         string  `json:"header"`
	HeaderStatic   string  `json:"header_static"`
	Locked         bool    `json:"locked"`
	FollowersCount int     `json:"followers_count"`
	FollowingCount int     `json:"following_count"`
	StatusesCount  int     `json:"statuses_count"`
	CreatedAt      string  `json:"created_at"`
	Bot            bool    `json:"bot"`
	Fields         []Field `json:"fields"`
}

type Field struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	VerifiedAt string `json:"verified_at"`
}

func (c *Client) GetAccount(accountID string) (*AccountFull, error) {
	respBody, err := c.doRequest("GET", "accounts/"+accountID, nil)
	if err != nil {
		return nil, err
	}

	var account AccountFull
	if err := json.Unmarshal(respBody, &account); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &account, nil
}

type Notification struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	CreatedAt string  `json:"created_at"`
	Account   Account `json:"account"`
	Status    *Status `json:"status"`
}

func (c *Client) GetNotifications(limit int, notificationType string) ([]Notification, error) {
	endpoint := "notifications"
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if notificationType != "" {
		query.Set("types", notificationType)
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	if err := json.Unmarshal(respBody, &notifications); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return notifications, nil
}

func (c *Client) GetAccountFollowers(accountID string, limit int) ([]Account, error) {
	endpoint := "accounts/" + accountID + "/followers"
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var accounts []Account
	if err := json.Unmarshal(respBody, &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return accounts, nil
}

func (c *Client) GetAccountFollowing(accountID string, limit int) ([]Account, error) {
	endpoint := "accounts/" + accountID + "/following"
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var accounts []Account
	if err := json.Unmarshal(respBody, &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return accounts, nil
}
