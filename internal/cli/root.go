package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	var useTUI bool

	cmd := &cobra.Command{
		Use:     "goreview",
		Short:   "Local-first code review tool for GitHub and GitLab",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if useTUI {
				fmt.Println("TUI mode not yet implemented")
				return nil
			}
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().BoolVar(&useTUI, "tui", false, "Launch interactive TUI mode")

	cmd.AddCommand(
		newAuthCmd(),
		newListCmd(),
		newReviewCmd(),
		newDiffCmd(),
		newCommentCmd(),
		newCommentsCmd(),
		newReviewsCmd(),
		newCheckoutCmd(),
		newPushCmd(),
	)

	return cmd
}
