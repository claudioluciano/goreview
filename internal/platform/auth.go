package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	GitHub string `yaml:"github,omitempty"`
	GitLab string `yaml:"gitlab,omitempty"`
}

func authConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "goreview", "auth.yaml"), nil
}

func LoadAuth() (*AuthConfig, error) {
	path, err := authConfigPath()
	if err != nil {
		return &AuthConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthConfig{}, nil
		}
		return nil, err
	}

	var cfg AuthConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveAuth(cfg *AuthConfig) error {
	path, err := authConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func ResolveToken(pt PlatformType) (string, error) {
	auth, err := LoadAuth()
	if err != nil {
		return "", err
	}

	// 1. Stored token
	switch pt {
	case GitHub:
		if auth.GitHub != "" {
			return auth.GitHub, nil
		}
	case GitLab:
		if auth.GitLab != "" {
			return auth.GitLab, nil
		}
	}

	// 2. CLI token
	switch pt {
	case GitHub:
		if token := cliToken("gh", "auth", "token"); token != "" {
			return token, nil
		}
	case GitLab:
		if token := cliToken("glab", "auth", "token"); token != "" {
			return token, nil
		}
	}

	// 3. Environment variable
	switch pt {
	case GitHub:
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			return token, nil
		}
	case GitLab:
		if token := os.Getenv("GITLAB_TOKEN"); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no auth token found for %s — run: goreview auth %s", pt, pt)
}

func cliToken(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
