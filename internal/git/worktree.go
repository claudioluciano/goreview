package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const worktreeDir = ".goreview/worktrees"

func WorktreePath(repoRoot, reviewID string) string {
	return filepath.Join(repoRoot, worktreeDir, reviewID)
}

func CreateWorktree(repoRoot, reviewID, branch string) (string, error) {
	path := WorktreePath(repoRoot, reviewID)

	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create worktree dir: %w", err)
	}

	// Fetch the branch first
	fetch := exec.Command("git", "-C", repoRoot, "fetch", "origin", branch)
	fetch.Stdout = os.Stdout
	fetch.Stderr = os.Stderr
	_ = fetch.Run() // best-effort, branch might be local

	cmd := exec.Command("git", "-C", repoRoot, "worktree", "add", path, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try with remote tracking branch
		cmd = exec.Command("git", "-C", repoRoot, "worktree", "add", path, "origin/"+branch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("create worktree for %s: %w", branch, err)
		}
	}

	return path, nil
}

func RemoveWorktree(repoRoot, reviewID string) error {
	path := WorktreePath(repoRoot, reviewID)

	cmd := exec.Command("git", "-C", repoRoot, "worktree", "remove", path, "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Fallback: remove directory and prune
		os.RemoveAll(path)
		prune := exec.Command("git", "-C", repoRoot, "worktree", "prune")
		prune.Run()
	}

	return nil
}
