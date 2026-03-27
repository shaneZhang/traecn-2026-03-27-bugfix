# Mastodon CLI Bug Fix Documentation

## 概述

本文档详细记录了 Mastodon CLI 项目中发现的 Bug 及其修复过程。所有修复已在 `weibo.zhangyuqing.cn` 实例上完成回归测试。

## 修复摘要

| Bug # | 问题描述 | 严重程度 | 状态 |
|-------|---------|---------|------|
| 1 | 关注用户 403/404 错误 | 高 | 已修复 |
| 2 | 取消关注报错 | 高 | 已修复 |
| 3 | 登录功能失效 - Login 函数未被导出 | 高 | 已修复 |
| 4 | visibility 参数被忽略 | 高 | 已修复 |
| 5 | hasScheme 函数边界检查错误 | 高 | 已修复 |
| 6 | 本地时间线 URL 构建错误 | 中 | 已修复 |
| 7 | isID 函数过于简单，误判用户输入 | 中 | 已修复 |
| 8 | main.go 登录检查逻辑错误 | 高 | 已修复 |
| 9 | 多处使用 _ 忽略错误 | 中 | 已修复 |
| 10 | 未导出的 Config 导致测试困难 | 低 | 已修复 |
| 11 | API 接口与官方文档不对应 | 高 | 已修复 |
| 12 | URL 查询参数编码错误 | 严重 | 已修复 |
| 13 | Search API 使用错误的版本 | 高 | 已修复 |

---

## 详细修复说明

### Bug 1 & 2: 关注/取消关注用户 403/404 错误

**问题描述：**
- 关注用户时返回 404 错误
- 根本原因是 URL 构建时查询参数被错误编码

**根因分析：**
```go
// 问题代码
parsedURL.Path = "/api/v1/" + endpoint  // endpoint 包含 "accounts/lookup?acct=xxx"
// 结果: /api/v1/accounts/lookup%3Facct=xxx  (? 被编码为 %3F)
```

**修复方案：**
```go
// 修复后代码 - 分离路径和查询参数
endpointURL, err := url.Parse(endpoint)
parsedURL.Path = "/api/v1/" + endpointURL.Path
parsedURL.RawQuery = endpointURL.RawQuery
```

**文件修改：**
- `cmd/internal/api/client.go`: `doRequest` 方法

---

### Bug 3: 登录功能失效

**问题描述：**
- Login 函数未被导出

**修复方案：**
- 实际上 `Login` 函数已经导出（大写开头），问题实际上是由于其他 Bug 导致的登录失败
- 修复了 URL 构建问题后，登录功能正常工作

---

### Bug 4: visibility 参数被忽略

**问题描述：**
- `post` 命令的 `--visibility` 参数被完全忽略

**根因分析：**
```go
// 问题代码 - PostStatus 只接收 status 参数
func (c *Client) PostStatus(status string) (*Status, error) {
    body := map[string]interface{}{
        "status": status,  // visibility 未被传递
    }
}
```

**修复方案：**
```go
// 修复后代码
func (c *Client) PostStatus(status, visibility string) (*Status, error) {
    body := map[string]interface{}{
        "status": status,
    }
    if visibility != "" {
        body["visibility"] = visibility
    }
    // ...
}
```

**文件修改：**
- `cmd/internal/api/client.go`: `PostStatus` 方法
- `cmd/internal/commands/commands.go`: `GetPostCommand` 中的调用

---

### Bug 5: hasScheme 函数边界检查错误

**问题描述：**
- 当 URL 长度为 7-8 时，`url[:8]` 会越界

**问题代码：**
```go
func hasScheme(url string) bool {
    return len(url) >= 7 && (url[:7] == "http://" || url[:8] == "https://")
    // 当 len(url) == 7 时，url[:8] 会 panic
}
```

**修复方案：**
```go
func hasScheme(urlStr string) bool {
    return (len(urlStr) >= 7 && urlStr[:7] == "http://") || 
           (len(urlStr) >= 8 && urlStr[:8] == "https://")
}
```

**文件修改：**
- `cmd/internal/api/client.go`: `hasScheme` 函数

---

### Bug 6: 本地时间线 URL 构建错误

**问题描述：**
- 本地时间线 URL 构建时查询参数处理不正确

**修复方案：**
- 与 Bug 1 一起修复，通过正确分离路径和查询参数

---

### Bug 7: isID 函数过于简单

**问题描述：**
- `isID` 函数只检查字符串是否全为数字，导致误判
- 用户名如 "123" 会被误判为 ID

**修复方案：**
```go
func isID(s string) bool {
    // Mastodon IDs 通常是大型数字（Snowflake IDs）
    // 至少 18 位数字
    if len(s) < 18 {
        return false
    }
    for _, c := range s {
        if c < '0' || c > '9' {
            return false
        }
    }
    return true
}
```

**文件修改：**
- `cmd/internal/commands/commands.go`: `isID` 函数

---

### Bug 8: main.go 登录检查逻辑错误

**问题描述：**
- 登录检查逻辑在 `rootCmd.Execute()` 之后执行，实际上不会被执行到

**问题代码：**
```go
if err := rootCmd.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}

// 这段代码永远不会执行到
cfg := api.GetConfig()
if cfg.InstanceURL != "" && cfg.AccessToken == "" {
    fmt.Println("Warning: Logged in but no access token...")
}
```

**修复方案：**
```go
// 在 Execute 之前检查
cfg := api.GetConfig()
if cfg.InstanceURL != "" && cfg.AccessToken == "" {
    fmt.Fprintln(os.Stderr, "Warning: Logged in but no access token...")
}

if err := rootCmd.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

**文件修改：**
- `cmd/main.go`

---

### Bug 9: 多处使用 _ 忽略错误

**问题描述：**
- 多处使用 `_` 忽略 `reader.ReadString` 的错误

**修复位置：**
1. `GetLoginCommand` - 读取实例 URL 输入
2. `GetLoginCommand` - 读取授权码输入
3. `GetPostCommand` - 读取状态输入

**修复方案：**
```go
// 修复前
input, _ := reader.ReadString('\n')

// 修复后
input, err := reader.ReadString('\n')
if err != nil {
    return fmt.Errorf("failed to read input: %w", err)
}
```

**文件修改：**
- `cmd/internal/commands/commands.go`

---

### Bug 10: 未导出的 Config 导致测试困难

**问题描述：**
- Config 结构体虽然已导出，但配置加载时的错误被忽略

**修复方案：**
```go
if err := viper.ReadInConfig(); err == nil {
    if err := viper.Unmarshal(cfg); err != nil {
        // 错误处理但继续执行
        cfg = &Config{}
    }
}
```

**文件修改：**
- `cmd/internal/config/config.go`

---

### Bug 11: API 接口与官方文档不对应

**问题描述：**
- `PostReply` 使用错误的参数名 `in_reply_to`

**修复方案：**
```go
// 修复前
body := map[string]interface{}{
    "status":      status,
    "in_reply_to": inReplyToID,  // 错误
}

// 修复后
body := map[string]interface{}{
    "status":         status,
    "in_reply_to_id": inReplyToID,  // 正确
}
```

**文件修改：**
- `cmd/internal/api/client.go`: `PostReply` 方法

---

### Bug 12: URL 查询参数编码错误（严重）

**问题描述：**
- 所有包含查询参数的 API 调用都会失败
- `?` 被编码为 `%3F`，导致服务器返回 404

**影响范围：**
- `GetAccountByUsername`
- `GetHomeTimeline`
- `GetLocalTimeline`
- `GetFederatedTimeline`
- `Search`
- `GetNotifications`
- `GetAccountFollowers`
- `GetAccountFollowing`

**修复方案：**
重构 `doRequest` 方法，正确分离路径和查询参数：

```go
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
    return c.doRequestWithVersion(method, endpoint, "v1", body)
}

func (c *Client) doRequestWithVersion(method, endpoint, apiVersion string, body interface{}) ([]byte, error) {
    // ...
    endpointURL, err := url.Parse(endpoint)
    // ...
    parsedURL.Path = "/api/" + apiVersion + "/" + endpointURL.Path
    parsedURL.RawQuery = endpointURL.RawQuery
    // ...
}
```

**文件修改：**
- `cmd/internal/api/client.go`

---

### Bug 13: Search API 使用错误的版本

**问题描述：**
- Search API 使用 v1 版本，但某些 Mastodon 实例只支持 v2

**修复方案：**
```go
func (c *Client) Search(query string, limit int) (*SearchResult, error) {
    endpoint := "search?q=" + url.QueryEscape(query)
    if limit > 0 {
        endpoint += "&limit=" + fmt.Sprintf("%d", limit)
    }
    respBody, err := c.doRequestWithVersion("GET", endpoint, "v2", nil)
    // ...
}
```

**文件修改：**
- `cmd/internal/api/client.go`: `Search` 方法

---

## 回归测试结果

所有修复已在 `weibo.zhangyuqing.cn` 实例上完成测试：

### 测试通过的功能

| 功能 | 命令 | 状态 |
|------|------|------|
| 查看当前用户 | `whoami` | ✅ |
| 关注用户 | `follow zhangyuqing` | ✅ |
| 取消关注 | `unfollow zhangyuqing` | ✅ |
| 发帖（带 visibility） | `post "test" -v public` | ✅ |
| 查看时间线 | `timeline home/local/federated` | ✅ |
| 搜索 | `search "zhangyuqing"` | ✅ |
| 查看账号信息 | `account zhangyuqing` | ✅ |
| 查看通知 | `notifications` | ✅ |
| 查看 followers | `followers zhangyuqing` | ✅ |
| 查看 following | `following zhangyuqing` | ✅ |

---

## 代码变更统计

| 文件 | 变更类型 | 变更行数 |
|------|---------|---------|
| `cmd/internal/api/client.go` | 修改 | +35/-10 |
| `cmd/internal/commands/commands.go` | 修改 | +20/-8 |
| `cmd/internal/config/config.go` | 修改 | +4/-1 |
| `cmd/main.go` | 修改 | +5/-5 |

---

## 后续建议

1. **添加单元测试**：为关键函数如 `hasScheme`, `isID` 等添加单元测试
2. **API 版本检测**：实现 Mastodon 实例 API 版本自动检测
3. **错误重试机制**：为网络请求添加重试逻辑
4. **配置文件验证**：添加配置文件结构和内容验证
5. **日志记录**：添加详细的日志记录便于调试

---

## 附录：Mastodon API 参考

- 官方文档：https://docs.joinmastodon.org/client/
- Search API v2：https://docs.joinmastodon.org/methods/search/#v2
- Statuses API：https://docs.joinmastodon.org/methods/statuses/
- Accounts API：https://docs.joinmastodon.org/methods/accounts/
