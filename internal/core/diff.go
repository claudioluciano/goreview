package core

type LineKind string

const (
	LineAdded   LineKind = "added"
	LineRemoved LineKind = "removed"
	LineContext LineKind = "context"
)

type Line struct {
	Kind    LineKind `yaml:"kind"`
	Content string   `yaml:"content"`
	OldNum  int      `yaml:"old_num,omitempty"`
	NewNum  int      `yaml:"new_num,omitempty"`
}

type Hunk struct {
	OldStart int    `yaml:"old_start"`
	OldCount int    `yaml:"old_count"`
	NewStart int    `yaml:"new_start"`
	NewCount int    `yaml:"new_count"`
	Header   string `yaml:"header"`
	Lines    []Line `yaml:"lines"`
}

type FileDiff struct {
	OldName   string `yaml:"old_name"`
	NewName   string `yaml:"new_name"`
	IsBinary  bool   `yaml:"is_binary"`
	IsNew     bool   `yaml:"is_new"`
	IsDeleted bool   `yaml:"is_deleted"`
	IsRenamed bool   `yaml:"is_renamed"`
	Hunks     []Hunk `yaml:"hunks"`
}

type DiffStat struct {
	File      string `yaml:"file"`
	Additions int    `yaml:"additions"`
	Deletions int    `yaml:"deletions"`
}
