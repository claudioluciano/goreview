package core

import "time"

type ReviewStatus string

const (
	ReviewDraft     ReviewStatus = "draft"
	ReviewPublished ReviewStatus = "published"
)

type FileStatus string

const (
	FileUnvisited FileStatus = "unvisited"
	FileReviewed  FileStatus = "reviewed"
	FileCommented FileStatus = "commented"
	FileSkipped   FileStatus = "skipped"
)

type CommentType string

const (
	CommentTypeComment  CommentType = "comment"
	CommentTypeBlocking CommentType = "blocking"
	CommentTypeNitpick  CommentType = "nitpick"
	CommentTypePraise   CommentType = "praise"
	CommentTypeQuestion CommentType = "question"
)

type Comment struct {
	ID        string      `yaml:"id"`
	File      string      `yaml:"file"`
	StartLine int         `yaml:"start_line"`
	EndLine   int         `yaml:"end_line"`
	Type      CommentType `yaml:"type"`
	Body      string      `yaml:"body"`
	CreatedAt time.Time   `yaml:"created_at"`
}

type Review struct {
	ID         string                `yaml:"id"`
	PR         int                   `yaml:"pr,omitempty"`
	Platform   string                `yaml:"platform,omitempty"`
	Repo       string                `yaml:"repo,omitempty"`
	Base       string                `yaml:"base"`
	Head       string                `yaml:"head"`
	CreatedAt  time.Time             `yaml:"created_at"`
	UpdatedAt  time.Time             `yaml:"updated_at"`
	Status     ReviewStatus          `yaml:"status"`
	FileStatus map[string]FileStatus `yaml:"file_status"`
	Comments   []Comment             `yaml:"comments"`
}

type Verdict string

const (
	VerdictApprove        Verdict = "approve"
	VerdictRequestChanges Verdict = "request-changes"
	VerdictComment        Verdict = "comment"
)

type ReviewSubmission struct {
	Verdict  Verdict   `yaml:"verdict"`
	Summary  string    `yaml:"summary,omitempty"`
	Comments []Comment `yaml:"comments"`
}

type PullRequest struct {
	Number    int       `yaml:"number"`
	Title     string    `yaml:"title"`
	Author    string    `yaml:"author"`
	Base      string    `yaml:"base"`
	Head      string    `yaml:"head"`
	Additions int       `yaml:"additions"`
	Deletions int       `yaml:"deletions"`
	Comments  int       `yaml:"comments"`
	Draft     bool      `yaml:"draft"`
	URL       string    `yaml:"url"`
	UpdatedAt time.Time `yaml:"updated_at"`
}
