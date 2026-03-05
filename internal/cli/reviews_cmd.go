package cli

import (
	"fmt"

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
				fmt.Println("No local reviews")
				return nil
			}

			for _, r := range reviews {
				prLabel := "local"
				if r.PR > 0 {
					prLabel = fmt.Sprintf("PR #%d", r.PR)
				}
				fmt.Printf("  %-30s %-10s %d comments  (%s)\n",
					r.ID, prLabel, len(r.Comments), r.Status)
			}

			return nil
		},
	}

	return cmd
}
