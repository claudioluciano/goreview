package cli

import (
	"encoding/json"
	"fmt"
	"os"

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
				fmt.Println(len(r.Comments))
				return nil
			}

			if jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(r.Comments)
			}

			if len(r.Comments) == 0 {
				fmt.Println("No comments yet")
				return nil
			}

			// Group by file
			grouped := make(map[string][]int)
			var order []string
			for i, c := range r.Comments {
				if _, seen := grouped[c.File]; !seen {
					order = append(order, c.File)
				}
				grouped[c.File] = append(grouped[c.File], i)
			}

			for _, file := range order {
				fmt.Printf("  %s\n", file)
				for _, i := range grouped[file] {
					c := r.Comments[i]
					lineSpec := fmt.Sprintf("L%d", c.StartLine)
					if c.StartLine != c.EndLine {
						lineSpec = fmt.Sprintf("L%d-%d", c.StartLine, c.EndLine)
					}
					typeLabel := ""
					if c.Type != "comment" {
						typeLabel = fmt.Sprintf("[%s] ", c.Type)
					}
					fmt.Printf("    %d. %s: %s%s\n", i+1, lineSpec, typeLabel, c.Body)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&count, "count", false, "Print only the comment count")

	return cmd
}
