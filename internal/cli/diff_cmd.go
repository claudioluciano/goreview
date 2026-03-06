package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/mattn/go-isatty"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/claudioluciano/goreview/internal/core"
	diffpkg "github.com/claudioluciano/goreview/internal/diff"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	var stat bool

	cmd := &cobra.Command{
		Use:   "diff [file]",
		Short: "Show diff for the active review",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			r, err := getActiveReview(app)
			if err != nil {
				return err
			}

			patch, err := app.repo.DiffTrees(r.Base, r.Head)
			if err != nil {
				return err
			}

			diffs := diffpkg.FromPatch(patch)

			if stat {
				return printStat(diffs)
			}

			if len(args) > 0 {
				fd := diffpkg.FilterFile(diffs, args[0])
				if fd == nil {
					return fmt.Errorf("file not found in diff: %s", args[0])
				}
				diffs = []core.FileDiff{*fd}
			}

			output := renderDiffs(diffs)

			if isatty.IsTerminal(os.Stdout.Fd()) {
				return pagerOutput(output)
			}

			lipgloss.Print(output)
			return nil
		},
	}

	cmd.Flags().BoolVar(&stat, "stat", false, "Show file change summary only")

	return cmd
}

func printStat(diffs []core.FileDiff) error {
	stats := diffpkg.Stat(diffs)
	totalAdd, totalDel := 0, 0

	lipgloss.Println()
	for _, s := range stats {
		name := styles.Bold.Render(fmt.Sprintf("%-50s", s.File))
		bar := styles.StatBar(s.Additions, s.Deletions)
		lipgloss.Printf("  %s %s\n", name, bar)
		totalAdd += s.Additions
		totalDel += s.Deletions
	}
	lipgloss.Println(styles.Separator())
	lipgloss.Printf("  %s  %s\n",
		styles.Faint.Render(fmt.Sprintf("%d files changed", len(stats))),
		styles.StatBar(totalAdd, totalDel))
	lipgloss.Println()
	return nil
}

func renderDiffs(diffs []core.FileDiff) string {
	var b strings.Builder
	for _, d := range diffs {
		name := diffpkg.FileName(d)

		label := ""
		switch {
		case d.IsNew:
			label = styles.Added.Render(" (new)")
		case d.IsDeleted:
			label = styles.Removed.Render(" (deleted)")
		case d.IsRenamed:
			label = styles.Warning.Render(fmt.Sprintf(" (renamed from %s)", d.OldName))
		}

		b.WriteString(styles.FileHdr.Render(" "+name+label) + "\n")

		if d.IsBinary {
			b.WriteString(styles.Faint.Render("  binary file") + "\n\n")
			continue
		}

		for _, h := range d.Hunks {
			b.WriteString(styles.HunkHdr.Render("  "+h.Header) + "\n")
			for _, l := range h.Lines {
				highlighted := diffpkg.HighlightLine(l, name)
				switch l.Kind {
				case core.LineAdded:
					lineNo := styles.LineNum.Render(fmt.Sprintf("%d", l.NewNum))
					gutter := styles.Added.Render("+")
					b.WriteString(fmt.Sprintf("  %s %s %s\n", lineNo, gutter, highlighted))
				case core.LineRemoved:
					lineNo := styles.LineNum.Render(fmt.Sprintf("%d", l.OldNum))
					gutter := styles.Removed.Render("-")
					b.WriteString(fmt.Sprintf("  %s %s %s\n", lineNo, gutter, highlighted))
				case core.LineContext:
					lineNo := styles.LineNum.Render(fmt.Sprintf("%d", l.NewNum))
					gutter := styles.Faint.Render(" ")
					b.WriteString(fmt.Sprintf("  %s %s %s\n", lineNo, gutter, highlighted))
				}
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func pagerOutput(content string) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	cmd := exec.Command(pager, "-R")
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stripColors(s string) string {
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

func getActiveReview(app *appContext) (*core.Review, error) {
	reviews, err := app.engine.List()
	if err != nil {
		return nil, err
	}

	if len(reviews) == 0 {
		return nil, fmt.Errorf("no active reviews — start one with: goreview review <target>")
	}

	latest := reviews[0]
	for _, r := range reviews[1:] {
		if r.UpdatedAt.After(latest.UpdatedAt) {
			latest = r
		}
	}
	return latest, nil
}
