package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repo struct {
	repo *git.Repository
	path string
}

func Open(path string) (*Repo, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}
	return &Repo{repo: r, path: path}, nil
}

func (r *Repo) Path() string {
	return r.path
}

func (r *Repo) ResolveRef(name string) (*plumbing.Hash, error) {
	// Try as branch
	ref, err := r.repo.Reference(plumbing.NewBranchReferenceName(name), true)
	if err == nil {
		h := ref.Hash()
		return &h, nil
	}

	// Try as remote branch
	ref, err = r.repo.Reference(plumbing.NewRemoteReferenceName("origin", name), true)
	if err == nil {
		h := ref.Hash()
		return &h, nil
	}

	// Try as tag
	ref, err = r.repo.Reference(plumbing.NewTagReferenceName(name), true)
	if err == nil {
		h := ref.Hash()
		return &h, nil
	}

	// Try as raw hash
	if plumbing.IsHash(name) {
		h := plumbing.NewHash(name)
		return &h, nil
	}

	// Try HEAD
	if name == "HEAD" {
		ref, err = r.repo.Head()
		if err == nil {
			h := ref.Hash()
			return &h, nil
		}
	}

	return nil, fmt.Errorf("cannot resolve ref: %s", name)
}

func (r *Repo) DiffTrees(baseRef, headRef string) (*object.Patch, error) {
	baseHash, err := r.ResolveRef(baseRef)
	if err != nil {
		return nil, err
	}
	headHash, err := r.ResolveRef(headRef)
	if err != nil {
		return nil, err
	}

	baseCommit, err := r.repo.CommitObject(*baseHash)
	if err != nil {
		return nil, fmt.Errorf("get base commit: %w", err)
	}
	headCommit, err := r.repo.CommitObject(*headHash)
	if err != nil {
		return nil, fmt.Errorf("get head commit: %w", err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get base tree: %w", err)
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get head tree: %w", err)
	}

	return baseTree.Patch(headTree)
}

func (r *Repo) RemoteURL(name string) (string, error) {
	remote, err := r.repo.Remote(name)
	if err != nil {
		return "", fmt.Errorf("get remote %s: %w", name, err)
	}
	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("remote %s has no URLs", name)
	}
	return urls[0], nil
}

func ParseRefRange(s string) (base, head string, ok bool) {
	parts := strings.SplitN(s, "..", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
