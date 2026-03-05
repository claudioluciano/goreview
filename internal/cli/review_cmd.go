package cli

import (
	"context"
	"fmt"
	"strconv"

	gitpkg "github.com/claudioluciano/goreview/internal/git"
	reviewpkg "github.com/claudioluciano/goreview/internal/review"
	"github.com/claudioluciano/goreview/internal/storage"
	"github.com/spf13/cobra"
)

func newReviewCmd() *cobra.Command {
	var (
		walk    bool
		resume  string
		discard string
		useTUI  bool
	)

	cmd := &cobra.Command{
		Use:   "review <target>",
		Short: "Start or resume a code review (PR number or branch range)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			if err := storage.EnsureGitignore(app.repo.Path()); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not update .gitignore: %v\n", err)
			}

			if discard != "" {
				return app.engine.Discard(discard)
			}

			if resume != "" {
				r, err := app.engine.Resume(resume)
				if err != nil {
					return err
				}
				fmt.Printf("Resumed review: %s (%d comments)\n", r.ID, len(r.Comments))
				return nil
			}

			if len(args) == 0 {
				return cmd.Help()
			}

			target := args[0]

			// Try as PR number
			if prNum, err := strconv.Atoi(target); err == nil {
				return startPRReview(app, prNum, walk, useTUI)
			}

			// Try as branch range
			base, head, ok := gitpkg.ParseRefRange(target)
			if !ok {
				return fmt.Errorf("invalid target: %s (use PR number or base..head)", target)
			}

			return startBranchReview(app, base, head, walk)
		},
	}

	cmd.Flags().BoolVar(&walk, "walk", false, "Guided file-by-file review mode")
	cmd.Flags().StringVar(&resume, "resume", "", "Resume an existing review by ID")
	cmd.Flags().StringVar(&discard, "discard", "", "Discard a review by ID")
	cmd.Flags().BoolVar(&useTUI, "tui", false, "Open review in TUI mode")

	return cmd
}

func startPRReview(app *appContext, prNum int, walk, tui bool) error {
	plat, pt, err := app.getPlatform()
	if err != nil {
		return err
	}

	pr, err := plat.GetPR(context.Background(), prNum)
	if err != nil {
		return err
	}

	id := reviewpkg.IDForPR(prNum)
	r, err := app.engine.Create(id, pr.Base, pr.Head, prNum, string(pt), "")
	if err != nil {
		return err
	}

	fmt.Printf("Review started: %s (PR #%d: %s)\n", r.ID, pr.Number, pr.Title)
	fmt.Printf("  %s → %s\n", pr.Head, pr.Base)

	if walk {
		return runWalkMode(app, r)
	}

	if tui {
		fmt.Println("TUI mode not yet implemented")
	}

	return nil
}

func startBranchReview(app *appContext, base, head string, walk bool) error {
	id := reviewpkg.IDForBranches(base, head)
	r, err := app.engine.Create(id, base, head, 0, "", "")
	if err != nil {
		return err
	}

	fmt.Printf("Review started: %s\n", r.ID)

	if walk {
		return runWalkMode(app, r)
	}

	return nil
}
