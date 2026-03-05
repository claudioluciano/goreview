package gitlab

import (
	"context"
	"fmt"

	gl "github.com/xanzy/go-gitlab"
	"github.com/claudioluciano/goreview/internal/core"
	"github.com/claudioluciano/goreview/internal/platform"
)

type Client struct {
	client *gl.Client
	owner  string
	repo   string
	pid    string
}

func New(token, owner, repo string) (*Client, error) {
	client, err := gl.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("create GitLab client: %w", err)
	}
	pid := fmt.Sprintf("%s/%s", owner, repo)
	return &Client{client: client, owner: owner, repo: repo, pid: pid}, nil
}

func (c *Client) ListPRs(ctx context.Context, opts platform.ListPRsOpts) ([]core.PullRequest, error) {
	state := "opened"
	glOpts := &gl.ListProjectMergeRequestsOptions{
		State:   &state,
		OrderBy: gl.Ptr("updated_at"),
		Sort:    gl.Ptr("desc"),
		ListOptions: gl.ListOptions{
			PerPage: 30,
		},
	}

	mrs, _, err := c.client.MergeRequests.ListProjectMergeRequests(c.pid, glOpts, gl.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("list MRs: %w", err)
	}

	var result []core.PullRequest
	for _, mr := range mrs {
		if opts.Author != "" && mr.Author.Username != opts.Author {
			continue
		}
		if mr.Draft && !opts.Draft {
			continue
		}
		result = append(result, core.PullRequest{
			Number:    mr.IID,
			Title:     mr.Title,
			Author:    mr.Author.Username,
			Base:      mr.TargetBranch,
			Head:      mr.SourceBranch,
			Draft:     mr.Draft,
			URL:       mr.WebURL,
			UpdatedAt: *mr.UpdatedAt,
		})
	}

	return result, nil
}

func (c *Client) GetPR(ctx context.Context, number int) (*core.PullRequest, error) {
	mr, _, err := c.client.MergeRequests.GetMergeRequest(c.pid, number, nil, gl.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("get MR !%d: %w", number, err)
	}

	return &core.PullRequest{
		Number:    mr.IID,
		Title:     mr.Title,
		Author:    mr.Author.Username,
		Base:      mr.TargetBranch,
		Head:      mr.SourceBranch,
		Draft:     mr.Draft,
		URL:       mr.WebURL,
		UpdatedAt: *mr.UpdatedAt,
	}, nil
}

func (c *Client) GetPRDiff(_ context.Context, _ int) ([]core.FileDiff, error) {
	return nil, fmt.Errorf("use local git diff instead")
}

func (c *Client) GetPRBranch(ctx context.Context, number int) (base, head string, err error) {
	mr, _, err := c.client.MergeRequests.GetMergeRequest(c.pid, number, nil, gl.WithContext(ctx))
	if err != nil {
		return "", "", fmt.Errorf("get MR !%d: %w", number, err)
	}
	return mr.TargetBranch, mr.SourceBranch, nil
}

func (c *Client) SubmitReview(ctx context.Context, number int, submission core.ReviewSubmission) error {
	// GitLab doesn't have a single "review" concept like GitHub.
	// We post individual discussion notes for each comment, then optionally approve.
	for _, cmt := range submission.Comments {
		body := formatCommentBody(cmt)
		posType := "text"
		opts := &gl.CreateMergeRequestDiscussionOptions{
			Body: &body,
			Position: &gl.PositionOptions{
				PositionType: &posType,
				NewPath:      &cmt.File,
				NewLine:      &cmt.StartLine,
			},
		}
		_, _, err := c.client.Discussions.CreateMergeRequestDiscussion(c.pid, number, opts, gl.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("create discussion on %s:%d: %w", cmt.File, cmt.StartLine, err)
		}
	}

	if submission.Verdict == core.VerdictApprove {
		_, _, err := c.client.MergeRequestApprovals.ApproveMergeRequest(c.pid, number, nil, gl.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("approve MR: %w", err)
		}
	}

	return nil
}

func formatCommentBody(c core.Comment) string {
	if c.Type == core.CommentTypeComment {
		return c.Body
	}
	return fmt.Sprintf("**[%s]** %s", c.Type, c.Body)
}
