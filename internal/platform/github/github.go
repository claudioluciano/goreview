package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v60/github"
	"github.com/claudioluciano/goreview/internal/core"
	"github.com/claudioluciano/goreview/internal/platform"
)

type Client struct {
	client *gh.Client
	owner  string
	repo   string
}

func New(token, owner, repo string) *Client {
	client := gh.NewClient(nil).WithAuthToken(token)
	return &Client{client: client, owner: owner, repo: repo}
}

func (c *Client) ListPRs(ctx context.Context, opts platform.ListPRsOpts) ([]core.PullRequest, error) {
	ghOpts := &gh.PullRequestListOptions{
		State:     "open",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: gh.ListOptions{
			PerPage: 30,
		},
	}

	prs, _, err := c.client.PullRequests.List(ctx, c.owner, c.repo, ghOpts)
	if err != nil {
		return nil, fmt.Errorf("list PRs: %w", err)
	}

	var result []core.PullRequest
	for _, pr := range prs {
		if opts.Author != "" && pr.GetUser().GetLogin() != opts.Author {
			continue
		}
		if pr.GetDraft() && !opts.Draft {
			continue
		}
		result = append(result, core.PullRequest{
			Number:    pr.GetNumber(),
			Title:     pr.GetTitle(),
			Author:    pr.GetUser().GetLogin(),
			Base:      pr.GetBase().GetRef(),
			Head:      pr.GetHead().GetRef(),
			Additions: pr.GetAdditions(),
			Deletions: pr.GetDeletions(),
			Comments:  pr.GetComments() + pr.GetReviewComments(),
			Draft:     pr.GetDraft(),
			URL:       pr.GetHTMLURL(),
			UpdatedAt: pr.GetUpdatedAt().Time,
		})
	}

	return result, nil
}

func (c *Client) GetPR(ctx context.Context, number int) (*core.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, number)
	if err != nil {
		return nil, fmt.Errorf("get PR #%d: %w", number, err)
	}

	return &core.PullRequest{
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		Author:    pr.GetUser().GetLogin(),
		Base:      pr.GetBase().GetRef(),
		Head:      pr.GetHead().GetRef(),
		Additions: pr.GetAdditions(),
		Deletions: pr.GetDeletions(),
		Comments:  pr.GetComments() + pr.GetReviewComments(),
		Draft:     pr.GetDraft(),
		URL:       pr.GetHTMLURL(),
		UpdatedAt: pr.GetUpdatedAt().Time,
	}, nil
}

func (c *Client) GetPRDiff(_ context.Context, _ int) ([]core.FileDiff, error) {
	// Diffs are computed locally via go-git, not from the API
	return nil, fmt.Errorf("use local git diff instead")
}

func (c *Client) GetPRBranch(ctx context.Context, number int) (base, head string, err error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, number)
	if err != nil {
		return "", "", fmt.Errorf("get PR #%d: %w", number, err)
	}
	return pr.GetBase().GetRef(), pr.GetHead().GetRef(), nil
}

func (c *Client) SubmitReview(ctx context.Context, number int, submission core.ReviewSubmission) error {
	event := "COMMENT"
	switch submission.Verdict {
	case core.VerdictApprove:
		event = "APPROVE"
	case core.VerdictRequestChanges:
		event = "REQUEST_CHANGES"
	}

	var comments []*gh.DraftReviewComment
	for _, cmt := range submission.Comments {
		side := "RIGHT"
		comment := &gh.DraftReviewComment{
			Path: &cmt.File,
			Body: gh.String(formatCommentBody(cmt)),
			Side: &side,
		}
		if cmt.StartLine != cmt.EndLine {
			comment.StartLine = &cmt.StartLine
			comment.Line = &cmt.EndLine
		} else {
			comment.Line = &cmt.StartLine
		}
		comments = append(comments, comment)
	}

	review := &gh.PullRequestReviewRequest{
		Event:    &event,
		Body:     &submission.Summary,
		Comments: comments,
	}

	_, _, err := c.client.PullRequests.CreateReview(ctx, c.owner, c.repo, number, review)
	if err != nil {
		return fmt.Errorf("submit review: %w", err)
	}

	return nil
}

func formatCommentBody(c core.Comment) string {
	if c.Type == core.CommentTypeComment {
		return c.Body
	}
	return fmt.Sprintf("**[%s]** %s", c.Type, c.Body)
}
