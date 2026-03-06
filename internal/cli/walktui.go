package cli

import (
	"fmt"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/claudioluciano/goreview/internal/core"
	diffpkg "github.com/claudioluciano/goreview/internal/diff"
	"github.com/claudioluciano/goreview/internal/review"
)

type walkMode int

const (
	modeNavigate walkMode = iota
	modeComment
)

type viewMode int

const (
	viewDiff viewMode = iota
	viewFull
)

type commentField int

const (
	fieldLineRange commentField = iota
	fieldType
	fieldBody
)

// cache key encodes file index + view mode
type cacheKey struct {
	idx  int
	view viewMode
}

type walkModel struct {
	app    *appContext
	engine *review.Engine
	review *core.Review
	diffs  []core.FileDiff
	stats  []core.DiffStat

	// file list
	fileIdx    int
	fileScroll int

	// diff viewport — cached per file+viewMode
	lineCache map[cacheKey][]string
	cacheMu   sync.RWMutex
	viewLines []string
	viewScroll int
	viewMode   viewMode

	// comment input
	mode         walkMode
	commentField commentField
	lineInput    string
	typeInput    string
	bodyInput    string

	// dimensions
	width  int
	height int

	// quit
	quitting bool
}

func newWalkModel(app *appContext, r *core.Review, diffs []core.FileDiff) walkModel {
	m := walkModel{
		app:       app,
		engine:    app.engine,
		review:    r,
		diffs:     diffs,
		stats:     diffpkg.Stat(diffs),
		lineCache: make(map[cacheKey][]string),
		width:     80,
		height:    24,
	}
	m.loadViewLines()
	return m
}

func runWalkTUI(app *appContext, r *core.Review) error {
	patch, err := app.repo.DiffTrees(r.Base, r.Head)
	if err != nil {
		return err
	}
	diffs := diffpkg.FromPatch(patch)
	if len(diffs) == 0 {
		fmt.Println("No changes found.")
		return nil
	}

	m := newWalkModel(app, r, diffs)
	p := tea.NewProgram(&m)
	_, err = p.Run()
	return err
}

type diffsCachedMsg struct{}

func (m *walkModel) Init() tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		for i := range m.diffs {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				lines := m.renderDiffLines(idx)
				m.cacheMu.Lock()
				m.lineCache[cacheKey{idx, viewDiff}] = lines
				m.cacheMu.Unlock()
			}(i)
		}
		wg.Wait()
		return diffsCachedMsg{}
	}
}

func (m *walkModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case diffsCachedMsg:
		m.loadViewLines()
		return m, nil

	case tea.KeyPressMsg:
		if m.mode == modeComment {
			return m.updateComment(msg)
		}
		return m.updateNavigate(msg)
	}
	return m, nil
}

func (m *walkModel) updateNavigate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	// --- file navigation ---
	case "j", "down":
		if m.fileIdx < len(m.diffs)-1 {
			m.fileIdx++
			m.viewScroll = 0
			m.loadViewLines()
			m.adjustFileScroll()
		}

	case "k", "up":
		if m.fileIdx > 0 {
			m.fileIdx--
			m.viewScroll = 0
			m.loadViewLines()
			m.adjustFileScroll()
		}

	case "tab", "n":
		if m.fileIdx < len(m.diffs)-1 {
			m.fileIdx++
			m.viewScroll = 0
			m.loadViewLines()
			m.adjustFileScroll()
		}

	case "shift+tab", "p":
		if m.fileIdx > 0 {
			m.fileIdx--
			m.viewScroll = 0
			m.loadViewLines()
			m.adjustFileScroll()
		}

	// --- diff scrolling ---
	case "J", "shift+down":
		m.scroll(3)

	case "K", "shift+up":
		m.scroll(-3)

	case "ctrl+d":
		m.scroll(m.contentHeight() / 2)

	case "ctrl+u":
		m.scroll(-m.contentHeight() / 2)

	case "g", "home":
		m.viewScroll = 0

	case "G", "end":
		m.viewScroll = m.maxScroll()

	// --- actions ---
	case "r":
		m.engine.SetFileStatus(m.review, m.currentFileName(), core.FileReviewed)

	case "s":
		m.engine.SetFileStatus(m.review, m.currentFileName(), core.FileSkipped)

	case "c":
		m.mode = modeComment
		m.commentField = fieldLineRange
		m.lineInput = ""
		m.typeInput = ""
		m.bodyInput = ""

	case "f":
		m.viewScroll = 0
		if m.viewMode == viewDiff {
			m.viewMode = viewFull
		} else {
			m.viewMode = viewDiff
		}
		m.loadViewLines()
	}

	return m, nil
}

func (m *walkModel) scroll(delta int) {
	m.viewScroll += delta
	if m.viewScroll < 0 {
		m.viewScroll = 0
	}
	if max := m.maxScroll(); m.viewScroll > max {
		m.viewScroll = max
	}
}

func (m *walkModel) updateComment(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.mode = modeNavigate
		return m, nil

	case "enter":
		switch m.commentField {
		case fieldLineRange:
			if m.lineInput == "" {
				return m, nil
			}
			m.commentField = fieldType
		case fieldType:
			if m.typeInput == "" {
				m.typeInput = "comment"
			}
			m.commentField = fieldBody
		case fieldBody:
			if m.bodyInput != "" {
				m.submitComment()
			}
			m.mode = modeNavigate
		}
		return m, nil

	case "backspace":
		switch m.commentField {
		case fieldLineRange:
			if len(m.lineInput) > 0 {
				m.lineInput = m.lineInput[:len(m.lineInput)-1]
			}
		case fieldType:
			if len(m.typeInput) > 0 {
				m.typeInput = m.typeInput[:len(m.typeInput)-1]
			}
		case fieldBody:
			if len(m.bodyInput) > 0 {
				m.bodyInput = m.bodyInput[:len(m.bodyInput)-1]
			}
		}
		return m, nil

	case "tab":
		switch m.commentField {
		case fieldLineRange:
			if m.lineInput != "" {
				m.commentField = fieldType
			}
		case fieldType:
			if m.typeInput == "" {
				m.typeInput = "comment"
			}
			m.commentField = fieldBody
		}
		return m, nil

	default:
		if len(key) == 1 {
			switch m.commentField {
			case fieldLineRange:
				m.lineInput += key
			case fieldType:
				m.typeInput += key
			case fieldBody:
				m.bodyInput += key
			}
		} else if key == "space" {
			switch m.commentField {
			case fieldBody:
				m.bodyInput += " "
			case fieldType:
				m.typeInput += " "
			}
		}
	}
	return m, nil
}

func (m *walkModel) submitComment() {
	file := m.currentFileName()
	startLine, endLine := parseLineRange(m.lineInput)
	if startLine == 0 {
		return
	}
	ctype := core.CommentType(m.typeInput)
	m.engine.AddComment(m.review, file, startLine, endLine, ctype, m.bodyInput)
}

func parseLineRange(s string) (start, end int) {
	if idx := strings.Index(s, "-"); idx != -1 {
		fmt.Sscanf(s[:idx], "%d", &start)
		fmt.Sscanf(s[idx+1:], "%d", &end)
	} else {
		fmt.Sscanf(s, "%d", &start)
		end = start
	}
	return
}

// ── View ──────────────────────────────────────────────────────────────

func (m *walkModel) View() tea.View {
	var content string
	if m.quitting {
		content = fmt.Sprintf("\n  %s %s\n\n",
			styles.Success.Render("Walk complete."),
			styles.Faint.Render(fmt.Sprintf("%d comments saved.", len(m.review.Comments))))
	} else {
		content = m.renderView()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeNone
	return v
}

func (m *walkModel) renderView() string {
	sidebarW := m.sidebarWidth()

	header := m.renderHeader()
	sidebar := m.renderSidebar(sidebarW)
	diffView := m.renderContentView()
	helpBar := m.renderHelpBar()

	// combine sidebar + content side by side
	sideLines := strings.Split(sidebar, "\n")
	contentLines := strings.Split(diffView, "\n")
	border := styles.Faint.Render("│")

	bodyHeight := m.contentHeight()

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	for i := 0; i < bodyHeight; i++ {
		sl := ""
		if i < len(sideLines) {
			sl = sideLines[i]
		}
		cl := ""
		if i < len(contentLines) {
			cl = contentLines[i]
		}
		b.WriteString(sl)
		b.WriteString(border)
		b.WriteString(cl)
		b.WriteString("\n")
	}

	b.WriteString(helpBar)
	return b.String()
}

// ── Header Bar ────────────────────────────────────────────────────────

func (m *walkModel) renderHeader() string {
	if m.fileIdx >= len(m.diffs) {
		return ""
	}

	d := m.diffs[m.fileIdx]
	name := m.currentFileName()

	// file label
	label := ""
	switch {
	case d.IsNew:
		label = styles.Added.Render(" new")
	case d.IsDeleted:
		label = styles.Removed.Render(" deleted")
	case d.IsRenamed:
		label = styles.Warning.Render(" renamed")
	}

	// stats
	stat := m.stats[m.fileIdx]
	statStr := ""
	if stat.Additions > 0 || stat.Deletions > 0 {
		parts := []string{}
		if stat.Additions > 0 {
			parts = append(parts, styles.Added.Render(fmt.Sprintf("+%d", stat.Additions)))
		}
		if stat.Deletions > 0 {
			parts = append(parts, styles.Removed.Render(fmt.Sprintf("-%d", stat.Deletions)))
		}
		statStr = " " + strings.Join(parts, " ")
	}

	// position
	pos := styles.Faint.Render(fmt.Sprintf(" %d/%d", m.fileIdx+1, len(m.diffs)))

	// view mode indicator
	modeStr := styles.Faint.Render(" [diff]")
	if m.viewMode == viewFull {
		modeStr = styles.Info.Render(" [full]")
	}

	// status
	status := m.review.FileStatus[name]
	statusStr := ""
	if status != "" && status != core.FileUnvisited {
		statusStr = " " + m.fileStatusBadge(status)
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Text).
		Background(styles.Subtle).
		Width(m.width)

	return headerStyle.Render(
		" " + styles.Bold.Render(name) + label + statStr + pos + modeStr + statusStr)
}

func (m *walkModel) fileStatusBadge(status core.FileStatus) string {
	switch status {
	case core.FileReviewed:
		return styles.Success.Render("✓ reviewed")
	case core.FileCommented:
		return styles.Warning.Render("● commented")
	case core.FileSkipped:
		return styles.Faint.Render("– skipped")
	default:
		return ""
	}
}

// ── Sidebar ───────────────────────────────────────────────────────────

func (m *walkModel) sidebarWidth() int {
	w := m.width / 4
	if w < 24 {
		w = 24
	}
	if w > 40 {
		w = 40
	}
	return w
}

func (m *walkModel) renderSidebar(width int) string {
	var b strings.Builder

	visibleHeight := m.contentHeight()
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	for i := m.fileScroll; i < len(m.diffs) && i-m.fileScroll < visibleHeight; i++ {
		d := m.diffs[i]
		name := diffpkg.FileName(d)
		status := m.review.FileStatus[name]
		stat := m.stats[i]

		icon := m.fileStatusIcon(status)

		// show just the filename, not full path, with dir prefix faded
		short := name
		maxNameW := width - 12 // icon + stats + padding
		if maxNameW < 8 {
			maxNameW = 8
		}
		if len(short) > maxNameW {
			short = "…" + short[len(short)-maxNameW+1:]
		}

		// mini stat
		miniStat := styles.Faint.Render("     ")
		if stat.Additions > 0 || stat.Deletions > 0 {
			miniStat = fmt.Sprintf("%s%s",
				styles.Added.Render(fmt.Sprintf("+%-2d", stat.Additions)),
				styles.Removed.Render(fmt.Sprintf("-%-2d", stat.Deletions)))
		}

		line := fmt.Sprintf(" %s %s %s", icon, miniStat, short)

		if i == m.fileIdx {
			sel := lipgloss.NewStyle().
				Background(styles.Subtle).
				Foreground(styles.Text).
				Bold(true).
				Width(width)
			b.WriteString(sel.Render(line))
		} else {
			norm := lipgloss.NewStyle().
				Foreground(styles.Subtext).
				Width(width)
			b.WriteString(norm.Render(line))
		}
		b.WriteString("\n")
	}

	// pad
	for i := len(m.diffs) - m.fileScroll; i < visibleHeight; i++ {
		b.WriteString(lipgloss.NewStyle().Width(width).Render(""))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m *walkModel) fileStatusIcon(status core.FileStatus) string {
	switch status {
	case core.FileReviewed:
		return styles.Success.Render("✓")
	case core.FileCommented:
		return styles.Warning.Render("●")
	case core.FileSkipped:
		return styles.Faint.Render("–")
	default:
		return styles.Faint.Render("○")
	}
}

// ── Content (diff or full file) ───────────────────────────────────────

func (m *walkModel) loadViewLines() {
	if m.fileIdx >= len(m.diffs) {
		m.viewLines = nil
		return
	}

	key := cacheKey{m.fileIdx, m.viewMode}
	m.cacheMu.RLock()
	cached, ok := m.lineCache[key]
	m.cacheMu.RUnlock()
	if ok {
		m.viewLines = cached
		return
	}

	var lines []string
	switch m.viewMode {
	case viewFull:
		lines = m.renderFullFileLines(m.fileIdx)
	default:
		lines = m.renderDiffLines(m.fileIdx)
	}

	m.cacheMu.Lock()
	m.lineCache[key] = lines
	m.cacheMu.Unlock()
	m.viewLines = lines
}

func (m *walkModel) renderDiffLines(idx int) []string {
	d := m.diffs[idx]
	name := diffpkg.FileName(d)
	var lines []string

	if d.IsBinary {
		lines = append(lines, styles.Faint.Render("  binary file"))
		return lines
	}

	addBg := lipgloss.NewStyle().Background(lipgloss.Color("#1a3a1a"))
	delBg := lipgloss.NewStyle().Background(lipgloss.Color("#3a1a1a"))
	foldStyle := styles.Faint

	for hi, h := range d.Hunks {
		// fold separator between hunks
		if hi > 0 {
			lines = append(lines, foldStyle.Render("  ⋯"))
		}

		lines = append(lines, styles.HunkHdr.Render("  "+h.Header))

		for _, l := range h.Lines {
			highlighted := diffpkg.HighlightLine(l, name)
			switch l.Kind {
			case core.LineAdded:
				lineNo := styles.LineNum.Render(fmt.Sprintf("%5d", l.NewNum))
				gutter := styles.Added.Render(" + ")
				content := addBg.Render(highlighted + " ")
				lines = append(lines, fmt.Sprintf(" %s%s%s", lineNo, gutter, content))
			case core.LineRemoved:
				lineNo := styles.LineNum.Render(fmt.Sprintf("%5d", l.OldNum))
				gutter := styles.Removed.Render(" - ")
				content := delBg.Render(highlighted + " ")
				lines = append(lines, fmt.Sprintf(" %s%s%s", lineNo, gutter, content))
			case core.LineContext:
				lineNo := styles.LineNum.Render(fmt.Sprintf("%5d", l.NewNum))
				gutter := styles.Faint.Render("   ")
				lines = append(lines, fmt.Sprintf(" %s%s%s", lineNo, gutter, highlighted))
			}
		}
	}

	return lines
}

func (m *walkModel) renderFullFileLines(idx int) []string {
	d := m.diffs[idx]
	name := diffpkg.FileName(d)

	if d.IsBinary {
		return []string{styles.Faint.Render("  binary file")}
	}
	if d.IsDeleted {
		return []string{styles.Removed.Render("  file deleted")}
	}

	// read head version of the file
	content, err := m.app.repo.FileContent(m.review.Head, name)
	if err != nil {
		return []string{styles.Error.Render("  " + err.Error())}
	}

	// build a set of changed line numbers from the diff
	addedLines := make(map[int]bool)
	for _, h := range d.Hunks {
		for _, l := range h.Lines {
			if l.Kind == core.LineAdded {
				addedLines[l.NewNum] = true
			}
		}
	}

	addBg := lipgloss.NewStyle().Background(lipgloss.Color("#1a3a1a"))
	fileLines := strings.Split(content, "\n")
	var lines []string

	for i, fl := range fileLines {
		num := i + 1
		lineNo := styles.LineNum.Render(fmt.Sprintf("%5d", num))

		highlighted := diffpkg.HighlightLine(core.Line{
			Kind: core.LineContext, Content: fl, NewNum: num,
		}, name)

		if addedLines[num] {
			gutter := styles.Added.Render(" + ")
			lines = append(lines, fmt.Sprintf(" %s%s%s", lineNo, gutter, addBg.Render(highlighted+" ")))
		} else {
			gutter := styles.Faint.Render("   ")
			lines = append(lines, fmt.Sprintf(" %s%s%s", lineNo, gutter, highlighted))
		}
	}

	return lines
}

func (m *walkModel) renderContentView() string {
	viewH := m.contentHeight()
	var b strings.Builder

	end := m.viewScroll + viewH
	if end > len(m.viewLines) {
		end = len(m.viewLines)
	}

	rendered := 0
	for i := m.viewScroll; i < end; i++ {
		b.WriteString(m.viewLines[i])
		b.WriteString("\n")
		rendered++
	}

	for i := rendered; i < viewH; i++ {
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m *walkModel) contentHeight() int {
	h := m.height - 3 // header + help + border
	if h < 1 {
		h = 1
	}
	return h
}

func (m *walkModel) maxScroll() int {
	max := len(m.viewLines) - m.contentHeight()
	if max < 0 {
		return 0
	}
	return max
}

// ── Help Bar ──────────────────────────────────────────────────────────

func (m *walkModel) renderHelpBar() string {
	if m.mode == modeComment {
		return m.renderCommentInput()
	}

	commentCount := len(m.review.Comments)

	// scroll position
	scrollStr := ""
	if max := m.maxScroll(); max > 0 {
		pct := (m.viewScroll * 100) / max
		scrollStr = styles.Faint.Render(fmt.Sprintf(" %d%%", pct))
	}

	status := styles.Faint.Render(fmt.Sprintf(" %d comments", commentCount))

	keys := []struct{ key, desc string }{
		{"j/k", "file"},
		{"J/K", "scroll"},
		{"^d/^u", "page"},
		{"f", "full file"},
		{"c", "comment"},
		{"r", "reviewed"},
		{"s", "skip"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts,
			styles.Bold.Render(k.key)+styles.Faint.Render(":"+k.desc))
	}

	help := strings.Join(parts, styles.Faint.Render(" │ "))
	return help + scrollStr + status
}

func (m *walkModel) renderCommentInput() string {
	file := m.currentFileName()
	prefix := styles.Info.Render(file) + " "

	var input string
	switch m.commentField {
	case fieldLineRange:
		input = prefix + styles.Prompt.Render("Line (e.g. 14 or 14-20): ") + m.lineInput + "█"
	case fieldType:
		input = prefix + styles.Prompt.Render("Type [comment/blocking/nitpick/praise/question]: ") + m.typeInput + "█"
	case fieldBody:
		input = prefix + styles.Prompt.Render("Comment: ") + m.bodyInput + "█"
	}

	esc := styles.Faint.Render(" (esc cancel, tab/enter next)")
	return input + esc
}

// ── Helpers ───────────────────────────────────────────────────────────

func (m *walkModel) currentFileName() string {
	if m.fileIdx >= len(m.diffs) {
		return ""
	}
	return diffpkg.FileName(m.diffs[m.fileIdx])
}

func (m *walkModel) adjustFileScroll() {
	vis := m.contentHeight()
	if vis < 1 {
		vis = 1
	}
	if m.fileIdx < m.fileScroll {
		m.fileScroll = m.fileIdx
	}
	if m.fileIdx >= m.fileScroll+vis {
		m.fileScroll = m.fileIdx - vis + 1
	}
}

func truncateStr(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return "…" + s[len(s)-maxLen+1:]
}
