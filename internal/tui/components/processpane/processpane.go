// Package processpane renders the right-pane variant for a process block:
// header + cmd / cwd / pid metadata + stderr tail card. Read-only — the
// process block exposes no controls in A1.
package processpane

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/data"
	"github.com/moq77111113/chop/internal/tui/ui"
)

const (
	cardLabel    = "STDERR · LAST 20"
	noOutputHint = "(no output yet)"
)

// Props bundles the identity, snapshot, and the static config strings the
// pane needs (cmd line + cwd) — pre-formatted by the caller so the
// component stays free of YAML / JSON decoding.
type Props struct {
	ID      string
	Type    string
	Snap    data.ProcessSnapshot
	HasSnap bool
	Cmd     string
	Cwd     string
}

// Styles bundles the palette: header, subtle, frame, label, stat key/value, pill.
type Styles struct {
	Header  lipgloss.Style
	Subtle  lipgloss.Style
	Frame   lipgloss.Style
	Label   lipgloss.Style
	StatKey lipgloss.Style
	StatVal lipgloss.Style
	Pill    ui.PillStyles
}

// Render produces the pane at width × height.
func Render(p Props, s Styles, width, _ int) string {
	header := renderHeader(p, s)
	meta := renderMeta(p, s)
	body := s.Label.Render(cardLabel) + "\n" + renderTail(p, s, max(width-4, 1))
	cardW := max(width-2, 1)
	card := s.Frame.Width(cardW).Render(body)
	return strings.Join([]string{header, "", meta, "", card}, "\n")
}

func renderHeader(p Props, s Styles) string {
	pill := ui.Pill(data.MapState(p.Snap.Status), s.Pill)
	return s.Header.Render(strings.ToUpper(p.ID)) + " " + s.Subtle.Render("("+p.Type+")") + "  " + pill
}

func renderMeta(p Props, s Styles) string {
	rows := []string{
		s.StatKey.Render("cmd  ") + s.StatVal.Render(p.Cmd),
	}
	if p.Cwd != "" {
		rows = append(rows, s.StatKey.Render("cwd  ")+s.StatVal.Render(p.Cwd))
	}
	rows = append(rows, s.StatKey.Render("pid  ")+s.StatVal.Render(pidOrExit(p)))
	return strings.Join(rows, "\n")
}

func pidOrExit(p Props) string {
	if !p.HasSnap {
		return "—"
	}
	if p.Snap.ExitCode != nil {
		return fmt.Sprintf("exit %d", *p.Snap.ExitCode)
	}
	if p.Snap.PID == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", p.Snap.PID)
}

func renderTail(p Props, s Styles, width int) string {
	if !p.HasSnap || len(p.Snap.StderrTail) == 0 {
		return s.Subtle.Render(noOutputHint)
	}
	lines := make([]string, len(p.Snap.StderrTail))
	for i, line := range p.Snap.StderrTail {
		lines[i] = s.Subtle.Render(truncate(line, width))
	}
	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}
