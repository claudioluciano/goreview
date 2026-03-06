package cli

import (
	"context"
	"fmt"
	"strconv"

	lipgloss "charm.land/lipgloss/v2"
	gitpkg "github.com/claudioluciano/goreview/internal/git"
	"github.com/claudioluciano/goreview/internal/cli/styles"
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
				lipgloss.Fprintf(cmd.ErrOrStderr(), "%s %v\n",
					styles.Warning.Render("warn:"), err)
			}

			if discard != "" {
				if err := app.engine.Discard(discard); err != nil {
					return err
				}
				lipgloss.Println(styles.Success.Render("Discarded review: " + discard))
				return nil
			}

			if resume != "" {
				r, err := app.engine.Resume(resume)
				if err != nil {
					return err
				}
				lipgloss.Printf("%s %s %s\n",
					styles.Success.Render("Resumed"),
					styles.Bold.Render(r.ID),
					styles.Faint.Render(fmt.Sprintf("(%d comments)", len(r.Comments))))
				return nil
			}

			if len(args) == 0 {
				return cmd.Help()
			}

			target := args[0]

			if prNum, err := strconv.Atoi(target); err == nil {
				return startPRReview(app, prNum, walk, useTUI)
			}

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

	lipgloss.Println()
	lipgloss.Printf("  %s %s\n",
		styles.Badge(fmt.Sprintf("PR #%d", pr.Number), styles.PRBadge),
		styles.Bold.Render(pr.Title))
	lipgloss.Printf("  %s %s %s\n",
		styles.Info.Render(pr.Head),
		styles.Faint.Render("->"),
		pr.Base)
	lipgloss.Printf("  %s  %s\n",
		styles.StatBar(pr.Additions, pr.Deletions),
		styles.Faint.Render("id:"+r.ID))
	lipgloss.Println()

	if walk {
		return runWalkTUI(app, r)
	}

	if tui {
		return runWalkTUI(app, r)
	}

	return nil
}

func startBranchReview(app *appContext, base, head string, walk bool) error {
	id := reviewpkg.IDForBranches(base, head)
	r, err := app.engine.Create(id, base, head, 0, "", "")
	if err != nil {
		return err
	}

	lipgloss.Println()
	lipgloss.Printf("  %s %s %s %s\n",
		styles.Success.Render("Review started"),
		styles.Info.Render(head),
		styles.Faint.Render("->"),
		base)
	lipgloss.Printf("  %s\n", styles.Faint.Render("id:"+r.ID))
	lipgloss.Println()

	if walk {
		return runWalkTUI(app, r)
	}

	return nil
}
