package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/claudioluciano/goreview/internal/core"
	"gopkg.in/yaml.v3"
)

const (
	goreviewDir = ".goreview"
	reviewsDir  = "reviews"
)

type Store struct {
	root string
}

func New(repoRoot string) *Store {
	return &Store{root: filepath.Join(repoRoot, goreviewDir)}
}

func (s *Store) ReviewsDir() string {
	return filepath.Join(s.root, reviewsDir)
}

func (s *Store) ensureDir() error {
	return os.MkdirAll(s.ReviewsDir(), 0o755)
}

func (s *Store) reviewPath(id string) string {
	safe := strings.ReplaceAll(id, "/", "_")
	return filepath.Join(s.ReviewsDir(), safe+".yaml")
}

func (s *Store) SaveReview(r *core.Review) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("create reviews dir: %w", err)
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal review: %w", err)
	}

	return os.WriteFile(s.reviewPath(r.ID), data, 0o644)
}

func (s *Store) LoadReview(id string) (*core.Review, error) {
	data, err := os.ReadFile(s.reviewPath(id))
	if err != nil {
		return nil, fmt.Errorf("read review %s: %w", id, err)
	}

	var r core.Review
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal review %s: %w", id, err)
	}

	return &r, nil
}

func (s *Store) ListReviews() ([]*core.Review, error) {
	entries, err := os.ReadDir(s.ReviewsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list reviews: %w", err)
	}

	var reviews []*core.Review
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".yaml")
		r, err := s.LoadReview(id)
		if err != nil {
			continue
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (s *Store) DeleteReview(id string) error {
	path := s.reviewPath(id)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete review %s: %w", id, err)
	}
	return nil
}
