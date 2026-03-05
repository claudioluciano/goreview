package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

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
		fmt.Printf("\n\033[1m── %s (%d/%d) ──\033[0m\n\n", name, i+1, len(diffs))

		for _, h := range d.Hunks {
			fmt.Printf("\033[36m%s\033[0m\n", h.Header)
			for _, l := range h.Lines {
				highlighted := diffpkg.HighlightLine(l, name)
				switch l.Kind {
				case core.LineAdded:
					fmt.Printf("\033[32m%4d+ %s\033[0m\n", l.NewNum, highlighted)
				case core.LineRemoved:
					fmt.Printf("\033[31m%4d- %s\033[0m\n", l.OldNum, highlighted)
				case core.LineContext:
					fmt.Printf("%4d  %s\n", l.NewNum, highlighted)
				}
			}
		}

		for {
			fmt.Print("\n> (l)ine comment, (s)kip file, (q)uit: ")
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
				fmt.Printf("\nWalk ended. %d comments saved.\n", len(r.Comments))
				return nil
			case "l":
				if err := walkComment(app, r, name, reader); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
				}
			default:
				fmt.Println("Invalid option. Use (l)ine comment, (s)kip, or (q)uit.")
			}
		}
	nextFile:
	}

	fmt.Printf("\nWalk complete. %d comments saved.\n", len(r.Comments))
	return nil
}

func walkComment(app *appContext, r *core.Review, file string, reader *bufio.Reader) error {
	fmt.Print("> Line or range (e.g. 14 or 14-20): ")
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

	fmt.Print("> Type (comment/blocking/nitpick/praise/question) [comment]: ")
	typeInput, _ := reader.ReadString('\n')
	typeInput = strings.TrimSpace(typeInput)
	if typeInput == "" {
		typeInput = "comment"
	}

	fmt.Print("> Comment: ")
	body, _ := reader.ReadString('\n')
	body = strings.TrimSpace(body)

	if body == "" {
		fmt.Println("Empty comment, skipped.")
		return nil
	}

	if err := app.engine.AddComment(r, file, startLine, endLine, core.CommentType(typeInput), body); err != nil {
		return err
	}

	fmt.Println("Added.")
	return nil
}
