package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/claudioluciano/goreview/internal/core"
	diffpkg "github.com/claudioluciano/goreview/internal/diff"
)

func runWalkMode(app *appContext, r *core.Review) error {
	patch, err := app.repo.DiffTrees(r.Base, r.Head)
	if err != nil {
		return err
	}

	diffs := diffpkg.FromPatch(patch)
	reader := bufio.NewReader(os.Stdin)

	for i, d := range diffs {
		name := diffpkg.FileName(d)

		lipgloss.Println()
		lipgloss.Printf("  %s %s\n",
			styles.FileHdr.Render(" "+name+" "),
			styles.FileNum.Render(fmt.Sprintf("(%d/%d)", i+1, len(diffs))))
		lipgloss.Println()

		for _, h := range d.Hunks {
			lipgloss.Printf("  %s\n", styles.HunkHdr.Render(h.Header))
			for _, l := range h.Lines {
				highlighted := diffpkg.HighlightLine(l, name)
				switch l.Kind {
				case core.LineAdded:
					lineNo := styles.LineNum.Render(fmt.Sprintf("%d", l.NewNum))
					gutter := styles.Added.Render("+")
					lipgloss.Printf("  %s %s %s\n", lineNo, gutter, highlighted)
				case core.LineRemoved:
					lineNo := styles.LineNum.Render(fmt.Sprintf("%d", l.OldNum))
					gutter := styles.Removed.Render("-")
					lipgloss.Printf("  %s %s %s\n", lineNo, gutter, highlighted)
				case core.LineContext:
					lineNo := styles.LineNum.Render(fmt.Sprintf("%d", l.NewNum))
					gutter := styles.Faint.Render(" ")
					lipgloss.Printf("  %s %s %s\n", lineNo, gutter, highlighted)
				}
			}
		}

		for {
			lipgloss.Printf("\n  %s ",
				styles.Prompt.Render("[l]ine comment  [s]kip  [q]uit >"))
			input, err := reader.ReadString('\n')
			if err != nil {
				return nil
			}
			input = strings.TrimSpace(input)

			switch input {
			case "s", "":
				app.engine.SetFileStatus(r, name, core.FileSkipped)
				goto nextFile
			case "q":
				lipgloss.Printf("\n  %s %s\n\n",
					styles.Success.Render("Walk ended."),
					styles.Faint.Render(fmt.Sprintf("%d comments saved.", len(r.Comments))))
				return nil
			case "l":
				if err := walkComment(app, r, name, reader); err != nil {
					lipgloss.Fprintln(os.Stderr, styles.Error.Render("error: "+err.Error()))
				}
			default:
				lipgloss.Println(styles.Warning.Render("  Use [l]ine comment, [s]kip, or [q]uit"))
			}
		}
	nextFile:
	}

	lipgloss.Printf("\n  %s %s\n\n",
		styles.Success.Render("Walk complete."),
		styles.Faint.Render(fmt.Sprintf("%d comments saved.", len(r.Comments))))
	return nil
}

func walkComment(app *appContext, r *core.Review, file string, reader *bufio.Reader) error {
	lipgloss.Printf("  %s ", styles.Prompt.Render("Line or range (e.g. 14 or 14-20) >"))
	lineInput, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	lineInput = strings.TrimSpace(lineInput)

	var startLine, endLine int
	if dashIdx := strings.Index(lineInput, "-"); dashIdx != -1 {
		startLine, err = strconv.Atoi(lineInput[:dashIdx])
		if err != nil {
			return fmt.Errorf("invalid start line: %s", lineInput[:dashIdx])
		}
		endLine, err = strconv.Atoi(lineInput[dashIdx+1:])
		if err != nil {
			return fmt.Errorf("invalid end line: %s", lineInput[dashIdx+1:])
		}
	} else {
		startLine, err = strconv.Atoi(lineInput)
		if err != nil {
			return fmt.Errorf("invalid line: %s", lineInput)
		}
		endLine = startLine
	}

	lipgloss.Printf("  %s ",
		styles.Prompt.Render("Type [comment/blocking/nitpick/praise/question] >"))
	typeInput, _ := reader.ReadString('\n')
	typeInput = strings.TrimSpace(typeInput)
	if typeInput == "" {
		typeInput = "comment"
	}

	lipgloss.Printf("  %s ", styles.Prompt.Render("Comment >"))
	body, _ := reader.ReadString('\n')
	body = strings.TrimSpace(body)

	if body == "" {
		lipgloss.Println(styles.Faint.Render("  Empty comment, skipped."))
		return nil
	}

	if err := app.engine.AddComment(r, file, startLine, endLine, core.CommentType(typeInput), body); err != nil {
		return err
	}

	lipgloss.Printf("  %s %s at %s\n",
		styles.Success.Render("Added"),
		styles.CommentTypeBadge(typeInput),
		styles.Info.Render(fmt.Sprintf("%s:%d", file, startLine)))
	return nil
}
