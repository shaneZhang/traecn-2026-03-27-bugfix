package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 存储 Mastodon CLI 的配置信息
type Config struct {
	InstanceURL  string `mapstructure:"instance_url"`
	AccessToken  string `mapstructure:"access_token"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

var cfg *Config

// GetConfig 获取当前配置
func GetConfig() *Config {
	if cfg != nil {
		return cfg
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	configDir := filepath.Join(homeDir, ".mastodon-cli")
	configFile := filepath.Join(configDir, "config.yaml")

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	cfg = &Config{}
	if err := viper.ReadInConfig(); err == nil {
		viper.Unmarshal(cfg)
	}

	return cfg
}

// SaveConfig 保存配置
func SaveConfig(c *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".mastodon-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	viper.Set("instance_url", c.InstanceURL)
	viper.Set("access_token", c.AccessToken)
	viper.Set("client_id", c.ClientID)
	viper.Set("client_secret", c.ClientSecret)

	return viper.WriteConfig()
}

// IsLoggedIn 检查是否已登录
func IsLoggedIn() bool {
	c := GetConfig()
	return c.InstanceURL != "" && c.AccessToken != ""
}

// ClearConfig 清除配置
func ClearConfig() error {
	cfg = &Config{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(homeDir, ".mastodon-cli", "config.yaml")
	return os.Remove(configFile)
}
