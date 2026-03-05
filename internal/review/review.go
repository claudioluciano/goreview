package review

import (
	"fmt"
	"time"

	"github.com/claudioluciano/goreview/internal/core"
	"github.com/claudioluciano/goreview/internal/storage"
)

type Engine struct {
	store *storage.Store
}

func New(store *storage.Store) *Engine {
	return &Engine{store: store}
}

func IDForPR(number int) string {
	return fmt.Sprintf("pr-%d", number)
}

func IDForBranches(base, head string) string {
	return fmt.Sprintf("%s..%s", base, head)
}

func (e *Engine) Create(id, base, head string, pr int, platform, repo string) (*core.Review, error) {
	existing, err := e.store.LoadReview(id)
	if err == nil && existing != nil {
		return existing, nil
	}

	now := time.Now()
	r := &core.Review{
		ID:         id,
		PR:         pr,
		Platform:   platform,
		Repo:       repo,
		Base:       base,
		Head:       head,
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     core.ReviewDraft,
		FileStatus: make(map[string]core.FileStatus),
		Comments:   []core.Comment{},
	}

	if err := e.store.SaveReview(r); err != nil {
		return nil, fmt.Errorf("save new review: %w", err)
	}

	return r, nil
}

func (e *Engine) Resume(id string) (*core.Review, error) {
	r, err := e.store.LoadReview(id)
	if err != nil {
		return nil, fmt.Errorf("resume review: %w", err)
	}
	return r, nil
}

func (e *Engine) List() ([]*core.Review, error) {
	return e.store.ListReviews()
}

func (e *Engine) Discard(id string) error {
	return e.store.DeleteReview(id)
}

func (e *Engine) AddComment(r *core.Review, file string, startLine, endLine int, ctype core.CommentType, body string) error {
	if startLine > endLine {
		return fmt.Errorf("invalid line range: %d > %d", startLine, endLine)
	}

	c := core.Comment{
		ID:        fmt.Sprintf("c%d", len(r.Comments)+1),
		File:      file,
		StartLine: startLine,
		EndLine:   endLine,
		Type:      ctype,
		Body:      body,
		CreatedAt: time.Now(),
	}

	r.Comments = append(r.Comments, c)
	r.FileStatus[file] = core.FileCommented
	r.UpdatedAt = time.Now()

	return e.store.SaveReview(r)
}

func (e *Engine) EditComment(r *core.Review, index int, body string) error {
	if index < 0 || index >= len(r.Comments) {
		return fmt.Errorf("comment index %d out of range (0-%d)", index, len(r.Comments)-1)
	}

	r.Comments[index].Body = body
	r.UpdatedAt = time.Now()

	return e.store.SaveReview(r)
}

func (e *Engine) DeleteComment(r *core.Review, index int) error {
	if index < 0 || index >= len(r.Comments) {
		return fmt.Errorf("comment index %d out of range (0-%d)", index, len(r.Comments)-1)
	}

	r.Comments = append(r.Comments[:index], r.Comments[index+1:]...)
	r.UpdatedAt = time.Now()

	return e.store.SaveReview(r)
}

func (e *Engine) SetFileStatus(r *core.Review, file string, status core.FileStatus) error {
	if r.FileStatus == nil {
		r.FileStatus = make(map[string]core.FileStatus)
	}
	r.FileStatus[file] = status
	r.UpdatedAt = time.Now()

	return e.store.SaveReview(r)
}

func (e *Engine) MarkPublished(r *core.Review) error {
	r.Status = core.ReviewPublished
	r.UpdatedAt = time.Now()
	return e.store.SaveReview(r)
}
