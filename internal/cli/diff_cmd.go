package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
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

			// Filter to single file if specified
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

			fmt.Print(stripColors(output))
			return nil
		},
	}

	cmd.Flags().BoolVar(&stat, "stat", false, "Show file change summary only")

	return cmd
}

func printStat(diffs []core.FileDiff) error {
	stats := diffpkg.Stat(diffs)
	totalAdd, totalDel := 0, 0
	for _, s := range stats {
		fmt.Printf("  %-50s | +%-4d -%d\n", s.File, s.Additions, s.Deletions)
		totalAdd += s.Additions
		totalDel += s.Deletions
	}
	fmt.Printf("  %d files changed, +%d -%d\n", len(stats), totalAdd, totalDel)
	return nil
}

func renderDiffs(diffs []core.FileDiff) string {
	var b strings.Builder
	for _, d := range diffs {
		name := diffpkg.FileName(d)
		b.WriteString(fmt.Sprintf("\033[1m── %s ──\033[0m\n", name))
		for _, h := range d.Hunks {
			b.WriteString(fmt.Sprintf("\033[36m%s\033[0m\n", h.Header))
			for _, l := range h.Lines {
				highlighted := diffpkg.HighlightLine(l, name)
				switch l.Kind {
				case core.LineAdded:
					b.WriteString(fmt.Sprintf("\033[32m%4d+ %s\033[0m\n", l.NewNum, highlighted))
				case core.LineRemoved:
					b.WriteString(fmt.Sprintf("\033[31m%4d- %s\033[0m\n", l.OldNum, highlighted))
				case core.LineContext:
					b.WriteString(fmt.Sprintf("%4d  %s\n", l.NewNum, highlighted))
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
	// Simple ANSI escape stripper
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// Skip until 'm'
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
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

	// Return most recently updated
	latest := reviews[0]
	for _, r := range reviews[1:] {
		if r.UpdatedAt.After(latest.UpdatedAt) {
			latest = r
		}
	}
	return latest, nil
}
