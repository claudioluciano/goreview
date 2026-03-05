package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/claudioluciano/goreview/internal/core"
	"github.com/spf13/cobra"
)

func newCommentCmd() *cobra.Command {
	var (
		commentType string
		edit        int
		delete      int
	)

	cmd := &cobra.Command{
		Use:   "comment <file>:<line> <body>",
		Short: "Add a comment to a file line or range",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext()
			if err != nil {
				return err
			}

			r, err := getActiveReview(app)
			if err != nil {
				return err
			}

			if edit >= 0 {
				if len(args) < 1 {
					return fmt.Errorf("usage: goreview comment --edit <n> <new body>")
				}
				return app.engine.EditComment(r, edit-1, strings.Join(args, " "))
			}

			if delete >= 0 {
				return app.engine.DeleteComment(r, delete-1)
			}

			if len(args) < 2 {
				return fmt.Errorf("usage: goreview comment <file>:<line[-endline]> <body>")
			}

			file, startLine, endLine, err := parseFileLineArg(args[0])
			if err != nil {
				return err
			}

			body := strings.Join(args[1:], " ")
			ct := core.CommentType(commentType)

			return app.engine.AddComment(r, file, startLine, endLine, ct, body)
		},
	}

	cmd.Flags().StringVarP(&commentType, "type", "t", "comment", "Comment type (comment, blocking, nitpick, praise, question)")
	cmd.Flags().IntVar(&edit, "edit", -1, "Edit comment by index (1-based)")
	cmd.Flags().IntVar(&delete, "delete", -1, "Delete comment by index (1-based)")

	return cmd
}

func parseFileLineArg(s string) (file string, startLine, endLine int, err error) {
	colonIdx := strings.LastIndex(s, ":")
	if colonIdx == -1 {
		return "", 0, 0, fmt.Errorf("expected format file:line or file:start-end, got: %s", s)
	}

	file = s[:colonIdx]
	lineSpec := s[colonIdx+1:]

	if dashIdx := strings.Index(lineSpec, "-"); dashIdx != -1 {
		startLine, err = strconv.Atoi(lineSpec[:dashIdx])
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid start line: %s", lineSpec[:dashIdx])
		}
		endLine, err = strconv.Atoi(lineSpec[dashIdx+1:])
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid end line: %s", lineSpec[dashIdx+1:])
		}
	} else {
		startLine, err = strconv.Atoi(lineSpec)
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid line number: %s", lineSpec)
		}
		endLine = startLine
	}

	return file, startLine, endLine, nil
}
