package diff

import (
	"fmt"
	"strings"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/claudioluciano/goreview/internal/core"
)

func FromPatch(patch *object.Patch) []core.FileDiff {
	var diffs []core.FileDiff

	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		fd := core.FileDiff{
			IsBinary: fp.IsBinary(),
		}

		switch {
		case from == nil && to != nil:
			fd.IsNew = true
			fd.NewName = to.Path()
		case from != nil && to == nil:
			fd.IsDeleted = true
			fd.OldName = from.Path()
		case from != nil && to != nil:
			fd.OldName = from.Path()
			fd.NewName = to.Path()
			if from.Path() != to.Path() {
				fd.IsRenamed = true
			}
		}

		if !fp.IsBinary() {
			fd.Hunks = buildHunks(fp.Chunks())
		}

		diffs = append(diffs, fd)
	}

	return diffs
}

func buildHunks(chunks []fdiff.Chunk) []core.Hunk {
	if len(chunks) == 0 {
		return nil
	}

	var hunks []core.Hunk
	oldLine := 1
	newLine := 1

	for _, chunk := range chunks {
		lines := splitLines(chunk.Content())
		if len(lines) == 0 {
			continue
		}

		hunk := core.Hunk{
			OldStart: oldLine,
			NewStart: newLine,
		}

		for _, line := range lines {
			switch chunk.Type() {
			case fdiff.Equal:
				hunk.Lines = append(hunk.Lines, core.Line{
					Kind:    core.LineContext,
					Content: line,
					OldNum:  oldLine,
					NewNum:  newLine,
				})
				oldLine++
				newLine++
				hunk.OldCount++
				hunk.NewCount++
			case fdiff.Add:
				hunk.Lines = append(hunk.Lines, core.Line{
					Kind:    core.LineAdded,
					Content: line,
					NewNum:  newLine,
				})
				newLine++
				hunk.NewCount++
			case fdiff.Delete:
				hunk.Lines = append(hunk.Lines, core.Line{
					Kind:    core.LineRemoved,
					Content: line,
					OldNum:  oldLine,
				})
				oldLine++
				hunk.OldCount++
			}
		}

		hunk.Header = fmt.Sprintf("@@ -%d,%d +%d,%d @@", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount)
		hunks = append(hunks, hunk)
	}

	return hunks
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func Stat(diffs []core.FileDiff) []core.DiffStat {
	var stats []core.DiffStat
	for _, d := range diffs {
		name := d.NewName
		if name == "" {
			name = d.OldName
		}
		s := core.DiffStat{File: name}
		for _, h := range d.Hunks {
			for _, l := range h.Lines {
				switch l.Kind {
				case core.LineAdded:
					s.Additions++
				case core.LineRemoved:
					s.Deletions++
				}
			}
		}
		stats = append(stats, s)
	}
	return stats
}

func FilterFile(diffs []core.FileDiff, path string) *core.FileDiff {
	for _, d := range diffs {
		if d.NewName == path || d.OldName == path {
			return &d
		}
	}
	return nil
}
