package styles

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

var (
	// Colors
	Green   = lipgloss.Color("#a6e3a1")
	Red     = lipgloss.Color("#f38ba8")
	Yellow  = lipgloss.Color("#f9e2af")
	Blue    = lipgloss.Color("#89b4fa")
	Cyan    = lipgloss.Color("#94e2d5")
	Magenta = lipgloss.Color("#cba6f7")
	Gray    = lipgloss.Color("#6c7086")
	Subtle  = lipgloss.Color("#45475a")
	Text    = lipgloss.Color("#cdd6f4")
	Subtext = lipgloss.Color("#a6adc8")

	// Styles
	Bold      = lipgloss.NewStyle().Bold(true)
	Faint     = lipgloss.NewStyle().Foreground(Gray)
	Header    = lipgloss.NewStyle().Bold(true).Foreground(Blue)
	Success   = lipgloss.NewStyle().Foreground(Green)
	Error     = lipgloss.NewStyle().Foreground(Red)
	Warning   = lipgloss.NewStyle().Foreground(Yellow)
	Info      = lipgloss.NewStyle().Foreground(Cyan)
	Highlight = lipgloss.NewStyle().Foreground(Magenta)

	// Diff styles
	Added   = lipgloss.NewStyle().Foreground(Green)
	Removed = lipgloss.NewStyle().Foreground(Red)
	Context = lipgloss.NewStyle().Foreground(Subtext)
	HunkHdr = lipgloss.NewStyle().Foreground(Cyan).Faint(true)
	FileHdr = lipgloss.NewStyle().Bold(true).Foreground(Blue).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Subtle)

	// Stat bar
	StatAdd = lipgloss.NewStyle().Foreground(Green)
	StatDel = lipgloss.NewStyle().Foreground(Red)

	// Badges
	badgeBase = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	DraftBadge    = badgeBase.Background(Yellow).Foreground(lipgloss.Color("#1e1e2e"))
	PRBadge       = badgeBase.Background(Blue).Foreground(lipgloss.Color("#1e1e2e"))
	ApprovedBadge = badgeBase.Background(Green).Foreground(lipgloss.Color("#1e1e2e"))
	BlockBadge    = badgeBase.Background(Red).Foreground(lipgloss.Color("#1e1e2e"))

	// Comment type badges
	CommentBadge  = lipgloss.NewStyle().Foreground(Subtext)
	BlockingBadge = lipgloss.NewStyle().Foreground(Red).Bold(true)
	NitpickBadge  = lipgloss.NewStyle().Foreground(Yellow)
	PraiseBadge   = lipgloss.NewStyle().Foreground(Green)
	QuestionBadge = lipgloss.NewStyle().Foreground(Cyan)

	// Walk mode
	Prompt   = lipgloss.NewStyle().Foreground(Magenta).Bold(true)
	FileNum  = lipgloss.NewStyle().Foreground(Gray)
	LineNum  = lipgloss.NewStyle().Foreground(Gray).Width(5).Align(lipgloss.Right)
	Gutter   = lipgloss.NewStyle().Foreground(Gray)
	Selected = lipgloss.NewStyle().Background(Subtle)
)

func Badge(label string, style lipgloss.Style) string {
	return style.Render(label)
}

func CommentTypeBadge(t string) string {
	switch t {
	case "blocking":
		return BlockingBadge.Render("blocking")
	case "nitpick":
		return NitpickBadge.Render("nitpick")
	case "praise":
		return PraiseBadge.Render("praise")
	case "question":
		return QuestionBadge.Render("question")
	default:
		return CommentBadge.Render("comment")
	}
}

func StatBar(add, del int) string {
	const maxWidth = 20
	total := add + del
	if total == 0 {
		return Faint.Render("(no changes)")
	}

	addW := (add * maxWidth) / total
	delW := maxWidth - addW
	if add > 0 && addW == 0 {
		addW = 1
		delW = maxWidth - 1
	}
	if del > 0 && delW == 0 {
		delW = 1
		addW = maxWidth - 1
	}

	bar := StatAdd.Render(strings.Repeat("+", addW)) +
		StatDel.Render(strings.Repeat("-", delW))

	return fmt.Sprintf("%s %s",
		bar,
		Faint.Render(fmt.Sprintf("+%d -%d", add, del)))
}

func StatusIcon(status string) string {
	switch status {
	case "draft":
		return Warning.Render("*")
	case "published":
		return Success.Render("*")
	default:
		return Faint.Render("*")
	}
}

func Separator() string {
	return Faint.Render(strings.Repeat("─", 60))
}
