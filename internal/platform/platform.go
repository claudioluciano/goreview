package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/claudioluciano/goreview/internal/core"
)

type ListPRsOpts struct {
	Author string
	Draft  bool
}

type Platform interface {
	ListPRs(ctx context.Context, opts ListPRsOpts) ([]core.PullRequest, error)
	GetPR(ctx context.Context, number int) (*core.PullRequest, error)
	GetPRDiff(ctx context.Context, number int) ([]core.FileDiff, error)
	SubmitReview(ctx context.Context, number int, submission core.ReviewSubmission) error
	GetPRBranch(ctx context.Context, number int) (base, head string, err error)
}

type PlatformType string

const (
	GitHub PlatformType = "github"
	GitLab PlatformType = "gitlab"
)

func DetectPlatform(remoteURL string) PlatformType {
	lower := strings.ToLower(remoteURL)
	switch {
	case strings.Contains(lower, "github.com"):
		return GitHub
	case strings.Contains(lower, "gitlab.com"), strings.Contains(lower, "gitlab"):
		return GitLab
	default:
		return GitHub
	}
}

func ParseRepoFromURL(remoteURL string) (owner, repo string, err error) {
	url := remoteURL
	url = strings.TrimSuffix(url, ".git")

	// SSH format: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH URL: %s", remoteURL)
		}
		path := parts[1]
		segments := strings.SplitN(path, "/", 2)
		if len(segments) != 2 {
			return "", "", fmt.Errorf("invalid repo path: %s", path)
		}
		return segments[0], segments[1], nil
	}

	// HTTPS format: https://github.com/owner/repo
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid URL: %s", remoteURL)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
