package diff

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/claudioluciano/goreview/internal/core"
)

func HighlightLine(line core.Line, filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return line.Content
	}

	iterator, err := lexer.Tokenise(nil, line.Content)
	if err != nil {
		return line.Content
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return line.Content
	}

	return strings.TrimRight(buf.String(), "\n")
}

func FileName(d core.FileDiff) string {
	if d.NewName != "" {
		return d.NewName
	}
	return d.OldName
}

func FileExt(d core.FileDiff) string {
	return filepath.Ext(FileName(d))
}
