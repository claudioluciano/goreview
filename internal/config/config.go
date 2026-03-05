package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Platform string `mapstructure:"platform"`
	Remote   string `mapstructure:"remote"`
}

func Load(repoRoot string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Global config
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "goreview"))
	}

	// Repo-level config (higher priority)
	v.AddConfigPath(filepath.Join(repoRoot, ".goreview"))

	v.SetDefault("platform", "")
	v.SetDefault("remote", "origin")

	// Ignore file-not-found errors — config is optional
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GlobalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "goreview")
	return dir, os.MkdirAll(dir, 0o755)
}
