package cli

import (
	"context"
	"fmt"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/claudioluciano/goreview/internal/cli/styles"
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
				lipgloss.Printf("%s Use %s, %s, or %s\n",
					styles.Warning.Render("No verdict specified."),
					styles.Bold.Render("--approve"),
					styles.Bold.Render("--request-changes"),
					styles.Bold.Render("--comment"))
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

			verdictStyle := styles.Success
			verdictLabel := string(verdict)
			switch verdict {
			case core.VerdictApprove:
				verdictStyle = styles.Success
				verdictLabel = "approved"
			case core.VerdictRequestChanges:
				verdictStyle = styles.Error
				verdictLabel = "changes requested"
			case core.VerdictComment:
				verdictStyle = styles.Info
				verdictLabel = "commented"
			}

			lipgloss.Printf("\n  %s %s %s\n\n",
				verdictStyle.Render(verdictLabel),
				styles.Faint.Render("on"),
				styles.Bold.Render(fmt.Sprintf("PR #%d", r.PR)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&approve, "approve", false, "Approve the PR")
	cmd.Flags().BoolVar(&requestChanges, "request-changes", false, "Request changes")
	cmd.Flags().BoolVar(&commentOnly, "comment", false, "Comment without verdict")
	cmd.Flags().StringVar(&summary, "summary", "", "Review summary message")

	return cmd
}
