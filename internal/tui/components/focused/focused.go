// Package focused renders the right-hand pane: link header, intent
// strip, four knobs (loss / latency / jitter / bandwidth), and a
// reset bar. Stateless — the focused-knob index lives on
// knobs.Pane and the link identity on the bubbletea adapter.
package focused

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/components/intentstrip"
	"github.com/moq77111113/chop/internal/tui/components/knob"
	"github.com/moq77111113/chop/internal/tui/intent"
	"github.com/moq77111113/chop/internal/tui/knobs"
	"github.com/moq77111113/chop/internal/tui/ui"
)

// Props carries the focused link identity, the knob pane snapshot, and
// the intent.Snapshot for the strip.
type Props struct {
	LinkID   string
	LinkType string
	Pane     *knobs.Pane
	Intent   intent.Snapshot
}

// Styles bundles header, subtle, knob palette, intentstrip palette,
// and the reset bar palette.
type Styles struct {
	Header lipgloss.Style
	Subtle lipgloss.Style
	Knob   knob.Styles
	Intent intentstrip.Styles
	Reset  ResetStyles
}

// ResetStyles is the bottom reset bar (frame + label + key chip).
type ResetStyles struct {
	Frame lipgloss.Style
	Label lipgloss.Style
	Kbd   ui.KbdStyles
}

// Render produces the pane at width × height.
func Render(p Props, s Styles, width, height int) string {
	if p.LinkID == "" {
		return s.Subtle.Render("no link selected")
	}
	header := s.Header.Render(strings.ToUpper(p.LinkID))
	if p.LinkType != "" {
		header += " " + s.Subtle.Render("("+p.LinkType+")")
	}
	parts := []string{
		header,
		"",
		intentstrip.Render(p.Intent, s.Intent, width),
		"",
	}
	for i := range knobs.KnobCount {
		k := p.Pane.Knob(knobs.KnobIndex(i))
		parts = append(parts, knob.Render(k, s.Knob, width, knobs.KnobIndex(i) == p.Pane.Focused()))
		parts = append(parts, "")
	}
	parts = append(parts, renderResetBar(s.Reset, width))
	out := strings.Join(parts, "\n")
	return clamp(out, height)
}

func renderResetBar(s ResetStyles, width int) string {
	frameW := max(width-2, 1) // Width(N)+Border = N+2 visible
	innerW := max(frameW-2, 1)
	label := s.Label.Render("undo all perturbations on this link")
	keys := ui.Kbd("r", s.Kbd) + " reset · " + ui.Kbd("R", s.Kbd) + " all"
	gap := max(innerW-lipgloss.Width(label)-lipgloss.Width(keys), 1)
	body := label + strings.Repeat(" ", gap) + keys
	return s.Frame.Width(frameW).Render(body)
}

func clamp(view string, height int) string {
	if height <= 0 {
		return view
	}
	lines := strings.Split(view, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
