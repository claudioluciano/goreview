package cli

import (
	"context"
	"fmt"
	"strconv"

	gitpkg "github.com/claudioluciano/goreview/internal/git"
	reviewpkg "github.com/claudioluciano/goreview/internal/review"
	"github.com/spf13/cobra"
)

func newCheckoutCmd() *cobra.Command {
	var clean bool

	cmd := &cobra.Command{
		Use:   "checkout <PR>",
		Short: "Checkout a PR branch into a git worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prNum, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("expected PR number, got: %s", args[0])
			}

			app, err := newAppContext()
			if err != nil {
				return err
			}

			reviewID := reviewpkg.IDForPR(prNum)

			if clean {
				gitpkg.RemoveWorktree(app.repo.Path(), reviewID)
				fmt.Printf("Removed worktree for PR #%d\n", prNum)
				return nil
			}

			plat, _, err := app.getPlatform()
			if err != nil {
				return err
			}

			_, head, err := plat.GetPRBranch(context.Background(), prNum)
			if err != nil {
				return err
			}

			path, err := gitpkg.CreateWorktree(app.repo.Path(), reviewID, head)
			if err != nil {
				return err
			}

			fmt.Printf("Worktree created at: %s\n", path)
			return nil
		},
	}

	cmd.Flags().BoolVar(&clean, "clean", false, "Remove the worktree instead of creating it")

	return cmd
}
