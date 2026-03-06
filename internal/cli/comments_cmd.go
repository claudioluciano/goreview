package cli

import (
	"encoding/json"
	"fmt"
	"os"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/spf13/cobra"
)

func newCommentsCmd() *cobra.Command {
	var (
		jsonFlag bool
		count    bool
	)

	cmd := &cobra.Command{
		Use:   "comments",
		Short: "List all comments in the active review",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			r, err := getActiveReview(app)
			if err != nil {
				return err
			}

			if count {
				lipgloss.Println(len(r.Comments))
				return nil
			}

			if jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(r.Comments)
			}

			if len(r.Comments) == 0 {
				lipgloss.Println(styles.Faint.Render("No comments yet"))
				return nil
			}

			grouped := make(map[string][]int)
			var order []string
			for i, c := range r.Comments {
				if _, seen := grouped[c.File]; !seen {
					order = append(order, c.File)
				}
				grouped[c.File] = append(grouped[c.File], i)
			}

			for _, file := range order {
				lipgloss.Printf("\n  %s\n", styles.Header.Render(file))
				for _, i := range grouped[file] {
					c := r.Comments[i]
					lineSpec := fmt.Sprintf("L%d", c.StartLine)
					if c.StartLine != c.EndLine {
						lineSpec = fmt.Sprintf("L%d-%d", c.StartLine, c.EndLine)
					}

					badge := styles.CommentTypeBadge(string(c.Type))
					lipgloss.Printf("    %s %s %s %s\n",
						styles.Faint.Render(fmt.Sprintf("%d.", i+1)),
						styles.Info.Render(lineSpec),
						badge,
						c.Body)
				}
			}
			lipgloss.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&count, "count", false, "Print only the comment count")

	return cmd
}
