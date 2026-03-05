package cli

import (
	"fmt"
	"os"

	"github.com/claudioluciano/goreview/internal/config"
	gitpkg "github.com/claudioluciano/goreview/internal/git"
	"github.com/claudioluciano/goreview/internal/platform"
	ghpkg "github.com/claudioluciano/goreview/internal/platform/github"
	glpkg "github.com/claudioluciano/goreview/internal/platform/gitlab"
	"github.com/claudioluciano/goreview/internal/review"
	"github.com/claudioluciano/goreview/internal/storage"
)

type appContext struct {
	repo   *gitpkg.Repo
	store  *storage.Store
	engine *review.Engine
	cfg    *config.Config
}

func newAppContext() (*appContext, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repo, err := gitpkg.Open(cwd)
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	cfg, err := config.Load(repo.Path())
	if err != nil {
		return nil, err
	}

	store := storage.New(repo.Path())
	engine := review.New(store)

	return &appContext{
		repo:   repo,
		store:  store,
		engine: engine,
		cfg:    cfg,
	}, nil
}

func (a *appContext) getPlatform() (platform.Platform, platform.PlatformType, error) {
	remoteURL, err := a.repo.RemoteURL(a.cfg.Remote)
	if err != nil {
		return nil, "", fmt.Errorf("get remote URL: %w", err)
	}

	pt := platform.DetectPlatform(remoteURL)
	if a.cfg.Platform != "" {
		pt = platform.PlatformType(a.cfg.Platform)
	}

	token, err := platform.ResolveToken(pt)
	if err != nil {
		return nil, "", err
	}

	owner, repo, err := platform.ParseRepoFromURL(remoteURL)
	if err != nil {
		return nil, "", err
	}

	switch pt {
	case platform.GitHub:
		return ghpkg.New(token, owner, repo), pt, nil
	case platform.GitLab:
		client, err := glpkg.New(token, owner, repo)
		if err != nil {
			return nil, "", err
		}
		return client, pt, nil
	default:
		return nil, "", fmt.Errorf("unsupported platform: %s", pt)
	}
}
