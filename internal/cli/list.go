package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/claudioluciano/goreview/internal/platform"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		author   string
		draft    bool
		jsonFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List open PRs/MRs for the current repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			plat, _, err := app.getPlatform()
			if err != nil {
				return err
			}

			prs, err := plat.ListPRs(context.Background(), platform.ListPRsOpts{
				Author: author,
				Draft:  draft,
			})
			if err != nil {
				return err
			}

			if jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(prs)
			}

			if len(prs) == 0 {
				lipgloss.Println(styles.Faint.Render("No open PRs found"))
				return nil
			}

			t := table.New().
				Headers("#", "TITLE", "BRANCH", "CHANGES", "STATUS").
				BorderStyle(lipgloss.NewStyle().Foreground(styles.Subtle)).
				StyleFunc(func(row, col int) lipgloss.Style {
					s := lipgloss.NewStyle().Padding(0, 1)
					if row == table.HeaderRow {
						return s.Bold(true).Foreground(styles.Blue)
					}
					return s.Foreground(styles.Text)
				})

			for _, pr := range prs {
				status := ""
				if pr.Draft {
					status = styles.Badge("draft", styles.DraftBadge)
				}
				comments := ""
				if pr.Comments > 0 {
					comments = styles.Faint.Render(fmt.Sprintf("%d comments", pr.Comments))
				}

				t.Row(
					fmt.Sprintf("%d", pr.Number),
					truncate(pr.Title, 40),
					styles.Info.Render(pr.Head)+styles.Faint.Render(" -> ")+pr.Base,
					styles.StatBar(pr.Additions, pr.Deletions),
					joinNonEmpty(" ", status, comments),
				)
			}

			lipgloss.Println(t)
			return nil
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "Filter by author")
	cmd.Flags().BoolVar(&draft, "draft", false, "Include draft PRs")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func joinNonEmpty(sep string, parts ...string) string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return ""
	}
	out := result[0]
	for _, p := range result[1:] {
		out += sep + p
	}
	return out
}
