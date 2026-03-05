package cli

import (
	"context"
	"fmt"

	"github.com/claudioluciano/goreview/internal/core"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	var (
		approve        bool
		requestChanges bool
		commentOnly    bool
		summary        string
	)

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Publish the active review's comments to the platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			r, err := getActiveReview(app)
			if err != nil {
				return err
			}

			if r.PR == 0 {
				return fmt.Errorf("review %s is not associated with a PR — cannot push", r.ID)
			}

			verdict := core.VerdictComment
			switch {
			case approve:
				verdict = core.VerdictApprove
			case requestChanges:
				verdict = core.VerdictRequestChanges
			case commentOnly:
				verdict = core.VerdictComment
			default:
				fmt.Println("No verdict specified. Use --approve, --request-changes, or --comment")
				return nil
			}

			plat, _, err := app.getPlatform()
			if err != nil {
				return err
			}

			submission := core.ReviewSubmission{
				Verdict:  verdict,
				Summary:  summary,
				Comments: r.Comments,
			}

			if err := plat.SubmitReview(context.Background(), r.PR, submission); err != nil {
				return fmt.Errorf("submit review: %w", err)
			}

			if err := app.engine.MarkPublished(r); err != nil {
				return fmt.Errorf("update local status: %w", err)
			}

			fmt.Printf("Review published for PR #%d (%s)\n", r.PR, verdict)
			return nil
		},
	}

	cmd.Flags().BoolVar(&approve, "approve", false, "Approve the PR")
	cmd.Flags().BoolVar(&requestChanges, "request-changes", false, "Request changes")
	cmd.Flags().BoolVar(&commentOnly, "comment", false, "Comment without verdict")
	cmd.Flags().StringVar(&summary, "summary", "", "Review summary message")

	return cmd
}
