package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// dimBehind faints the body string so a layered modal pops over it.
// Lipgloss's Faint flag emits the SGR 2 escape — terminals that
// honour it draw the body at half intensity; terminals that ignore it
// just show the body at full saturation, which is still readable.
func dimBehind(body string) string {
	return lipgloss.NewStyle().Faint(true).Render(body)
}

// overlayCenter splices modal rows over body rows so the modal reads
// as a centered overlay rather than replacing the entire frame. Both
// inputs are line-split; the modal's non-empty lines (lipgloss.Place
// pads the rest with spaces) substitute into the matching body lines.
func overlayCenter(body, modal string) string {
	bodyLines := strings.Split(body, "\n")
	modalLines := strings.Split(modal, "\n")
	out := make([]string, len(bodyLines))
	copy(out, bodyLines)
	for i := 0; i < len(modalLines) && i < len(out); i++ {
		ml := modalLines[i]
		if strings.TrimSpace(lipgloss.NewStyle().Render(stripANSI(ml))) == "" {
			continue
		}
		out[i] = ml
	}
	return strings.Join(out, "\n")
}

// stripANSI removes SGR escape sequences so emptiness checks see the
// printable content only. Cheap state machine — not a full ANSI parser.
func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			b.WriteRune(r)
		}
	}
	return b.String()
}
