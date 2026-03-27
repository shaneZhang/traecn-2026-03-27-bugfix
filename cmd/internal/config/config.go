package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	InstanceURL  string `mapstructure:"instance_url"`
	AccessToken  string `mapstructure:"access_token"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

var cfg *Config

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
		if err := viper.Unmarshal(cfg); err != nil {
			// Log error but continue with empty config
			cfg = &Config{}
		}
	}

	return cfg
}

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

func IsLoggedIn() bool {
	c := GetConfig()
	return c.InstanceURL != "" && c.AccessToken != ""
}

func ClearConfig() error {
	cfg = &Config{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(homeDir, ".mastodon-cli", "config.yaml")
	return os.Remove(configFile)
}
