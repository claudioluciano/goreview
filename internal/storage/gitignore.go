package storage

import (
	"os"
	"path/filepath"
	"strings"
)

func EnsureGitignore(repoRoot string) error {
	gitignorePath := filepath.Join(repoRoot, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == ".goreview/" {
			return nil
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry := ".goreview/\n"
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		entry = "\n" + entry
	}

	_, err = f.WriteString(entry)
	return err
}
