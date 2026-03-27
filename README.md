# Mastodon CLI

Mastodon 命令行终端助手 - 使用 Go 和 Cobra 开发的 Mastodon 客户端工具。

## 功能特性

- 🔐 **登录/退出** - OAuth2 认证方式登录 Mastodon 实例
- 📝 **发帖** - 发布状态 (toot) 到 Mastodon
- 👥 **关注/取消关注** - 管理关注关系
- 👤 **查看当前用户** - 显示已登录用户信息

## 技术栈

- [Go](https://go.dev/) - 编程语言
- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Mastodon API](https://docs.joinmastodon.org/client/) - REST API

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/yourusername/mastodon-cli.git
cd mastodon-cli

# 构建
go build -o mastodon-cli ./cmd

# 安装到系统 PATH
sudo cp mastodon-cli /usr/local/bin/
```

### 使用 Go install

```bash
go install github.com/yourusername/mastodon-cli@latest
```

## 使用方法

### 1. 登录

首次使用需要登录到 Mastodon 实例：

```bash
# 方式一：直接指定实例
mastodon-cli login mastodon.social

# 方式二：交互式输入
mastodon-cli login

# 方式三：使用标志
mastodon-cli login --instance mastodon.social
```

登录流程：
1. 程序会在指定 Mastodon 实例注册应用
2. 自动打开浏览器授权页面（或显示 URL 手动打开）
3. 输入授权码完成认证
4. 凭证保存在 `~/.mastodon-cli/config.yaml`

### 2. 查看当前登录状态

```bash
mastodon-cli whoami
```

输出示例：
```
Logged in as: @username@mastodon.social
Display name: Your Name
Account ID: 1234567890
```

### 3. 发布状态

```bash
# 直接发布
mastodon-cli post "Hello, Mastodon!"

# 从标准输入读取
echo "Hello from CLI!" | mastodon-cli post

# 交互式输入
mastodon-cli post
```

### 4. 关注用户

```bash
# 使用用户名（不含 @）
mastodon-cli follow username

# 或使用完整账号地址
mastodon-cli follow username@mastodon.social
```

### 5. 取消关注

```bash
mastodon-cli unfollow username
```

### 6. 退出登录

```bash
mastodon-cli logout
```

## 命令列表

| 命令 | 描述 |
|------|------|
| `mastodon-cli login [instance]` | 登录到 Mastodon 实例 |
| `mastodon-cli logout` | 退出登录 |
| `mastodon-cli post [status]` | 发布状态 |
| `mastodon-cli follow <username>` | 关注用户 |
| `mastodon-cli unfollow <username>` | 取消关注用户 |
| `mastodon-cli whoami` | 显示当前登录用户 |

## 配置说明

登录后，配置文件保存在 `~/.mastodon-cli/config.yaml`：

```yaml
instance_url: https://mastodon.social
access_token: your_access_token
client_id: your_client_id
client_secret: your_client_secret
```

## 开发相关

### 项目结构

```
mastodon-cli/
├── cmd/
│   ├── main.go                 # 主入口
│   └── internal/
│       ├── api/
│       │   └── client.go       # Mastodon API 客户端
│       ├── commands/
│       │   └── commands.go     # CLI 命令
│       └── config/
│           └── config.go       # 配置管理
├── go.mod
└── README.md
```

### 运行测试

```bash
go test ./...
```

### 代码规范检查

```bash
go fmt ./...
go vet ./...
```

## 许可证

MIT License
