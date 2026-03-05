package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
				fmt.Println("No open PRs found")
				return nil
			}

			for _, pr := range prs {
				draftLabel := ""
				if pr.Draft {
					draftLabel = " (draft)"
				}
				fmt.Printf("  #%-4d %-40s %s → %s  +%d -%d  %d comments%s\n",
					pr.Number, pr.Title, pr.Head, pr.Base,
					pr.Additions, pr.Deletions, pr.Comments, draftLabel)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "Filter by author")
	cmd.Flags().BoolVar(&draft, "draft", false, "Include draft PRs")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")

	return cmd
}
