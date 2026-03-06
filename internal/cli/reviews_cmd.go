package cli

import (
	"fmt"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/spf13/cobra"
)

func newReviewsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviews",
		Short: "List all local review sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			reviews, err := app.engine.List()
			if err != nil {
				return err
			}

			if len(reviews) == 0 {
				lipgloss.Println(styles.Faint.Render("No local reviews"))
				return nil
			}

			t := table.New().
				Headers("", "REVIEW", "TARGET", "COMMENTS", "STATUS").
				BorderStyle(lipgloss.NewStyle().Foreground(styles.Subtle)).
				StyleFunc(func(row, col int) lipgloss.Style {
					s := lipgloss.NewStyle().Padding(0, 1)
					if row == table.HeaderRow {
						return s.Bold(true).Foreground(styles.Blue)
					}
					return s.Foreground(styles.Text)
				})

			for _, r := range reviews {
				target := "local"
				if r.PR > 0 {
					target = styles.Badge(fmt.Sprintf("PR #%d", r.PR), styles.PRBadge)
				}

				status := styles.Faint.Render(string(r.Status))
				if r.Status == "published" {
					status = styles.Success.Render("published")
				}

				t.Row(
					styles.StatusIcon(string(r.Status)),
					styles.Bold.Render(r.ID),
					target,
					fmt.Sprintf("%d", len(r.Comments)),
					status,
				)
			}

			lipgloss.Println(t)
			return nil
		},
	}

	return cmd
}
