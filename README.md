# Mastodon CLI

Mastodon 命令行终端助手 - 使用 Go 和 Cobra 开发的 Mastodon 客户端工具。

## 功能特性

- 🔐 **登录/退出** - OAuth2 认证方式登录 Mastodon 实例
- 📝 **发帖** - 发布状态 (toot) 到 Mastodon
- 👥 **关注/取消关注** - 管理关注关系
- 👤 **查看当前用户** - 显示已登录用户信息
- 📰 **时间线** - 查看主页、本地、联邦时间线
- ❤️ **点赞/取消点赞** - 收藏或取消收藏帖子
- 🔁 **转推/取消转推** - 转发或取消转发帖子
- 💬 **回复** - 回复帖子
- 🗑️ **删除** - 删除自己的帖子
- 🔍 **搜索** - 搜索用户、帖子、话题标签
- 👀 **查看用户资料** - 查看用户详细信息、粉丝数、关注数
- 🔔 **通知** - 查看通知（提及、点赞、关注、转发等）
- 📋 **粉丝/关注列表** - 查看用户的粉丝和关注列表

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

### 6. 查看时间线

```bash
# 查看主页时间线（默认）
mastodon-cli timeline

# 查看主页时间线
mastodon-cli timeline home

# 查看本地时间线
mastodon-cli timeline local

# 查看联邦时间线
mastodon-cli timeline federated

# 指定获取数量
mastodon-cli timeline home -n 50
```

### 7. 查看帖子详情

```bash
mastodon-cli status <status_id>
```

### 8. 点赞/取消点赞

```bash
# 点赞帖子
mastodon-cli favourite <status_id>

# 取消点赞
mastodon-cli unfavourite <status_id>
```

### 9. 转推/取消转推

```bash
# 转推帖子
mastodon-cli boost <status_id>

# 取消转推
mastodon-cli unboost <status_id>
```

### 10. 回复帖子

```bash
mastodon-cli reply <status_id> <message>
```

### 11. 删除帖子

```bash
mastodon-cli delete <status_id>
```

### 12. 搜索

```bash
# 搜索用户、帖子、话题标签
mastodon-cli search "golang"

# 指定返回数量
mastodon-cli search "mastodon" -n 20
```

### 13. 查看用户资料

```bash
# 使用用户名查看
mastodon-cli account username

# 使用用户 ID 查看
mastodon-cli account <account_id>
```

### 14. 查看通知

```bash
# 查看所有通知
mastodon-cli notifications

# 指定获取数量
mastodon-cli notifications -n 10

# 按类型筛选 (mention, favourite, reblog, follow)
mastodon-cli notifications --type mention
```

### 15. 查看粉丝/关注列表

```bash
# 查看粉丝列表
mastodon-cli followers <username>

# 查看关注列表
mastodon-cli following <username>

# 指定获取数量
mastodon-cli followers username -n 50
```

### 16. 退出登录

```bash
mastodon-cli logout
```

## 命令列表

### 账户管理

| 命令 | 描述 |
|------|------|
| `mastodon-cli login [instance]` | 登录到 Mastodon 实例 |
| `mastodon-cli logout` | 退出登录 |
| `mastodon-cli whoami` | 显示当前登录用户 |

### 发帖互动

| 命令 | 描述 |
|------|------|
| `mastodon-cli post [status]` | 发布状态 |
| `mastodon-cli reply <status_id> <message>` | 回复帖子 |
| `mastodon-cli delete <status_id>` | 删除自己的帖子 |

### 帖子操作

| 命令 | 描述 |
|------|------|
| `mastodon-cli timeline [home\|local\|federated]` | 查看时间线 |
| `mastodon-cli status <status_id>` | 查看帖子详情 |
| `mastodon-cli favourite <status_id>` | 点赞帖子 |
| `mastodon-cli unfavourite <status_id>` | 取消点赞 |
| `mastodon-cli boost <status_id>` | 转推帖子 |
| `mastodon-cli unboost <status_id>` | 取消转推 |

### 关注关系

| 命令 | 描述 |
|------|------|
| `mastodon-cli follow <username>` | 关注用户 |
| `mastodon-cli unfollow <username>` | 取消关注 |
| `mastodon-cli followers <username>` | 查看粉丝列表 |
| `mastodon-cli following <username>` | 查看关注列表 |

### 搜索与发现

| 命令 | 描述 |
|------|------|
| `mastodon-cli search <query>` | 搜索用户、帖子、话题标签 |
| `mastodon-cli account <username\|id>` | 查看用户资料 |
| `mastodon-cli notifications` | 查看通知 |

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
