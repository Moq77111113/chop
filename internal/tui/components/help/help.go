// Package help renders the keymap modal opened with ?. The card is
// centered over the body; the bubbletea adapter handles the body dim.
package help

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/ui"
)

const (
	cardWidth   = 72
	keyColWidth = 20 // fits up to 5 single-char chips with one glue token
	leftIndent  = "  "
	footerHint  = "? or esc to close"
)

// Group is a labeled cluster of bindings shown on one row of the card.
type Group struct {
	Title    string
	Bindings []Binding
}

// Binding is a (keys, description) pair shown in the help card.
type Binding struct {
	Keys string
	Desc string
}

// Props carries the binding groups + version stamp for the subtitle.
type Props struct {
	Groups  []Group
	Version string // "v0.4.1-dev" or "dev"
}

// Styles bundles the modal palette.
type Styles struct {
	Frame  lipgloss.Style
	Title  lipgloss.Style
	IPA    lipgloss.Style
	Group  lipgloss.Style
	Desc   lipgloss.Style
	Subtle lipgloss.Style
	Kbd    ui.KbdStyles
}

// Render returns the modal centered at width × height.
func Render(p Props, s Styles, width, height int) string {
	body := bodyLines(p, s)
	card := s.Frame.Width(min(cardWidth, width-2)).Render(body)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
}

func bodyLines(p Props, s Styles) string {
	var lines []string
	lines = append(lines, s.Title.Render("chop · keymap"))
	if p.Version != "" {
		lines = append(lines, s.IPA.Render("/tʃɒp/ · "+p.Version))
	}
	for i, g := range p.Groups {
		if i == 0 {
			lines = append(lines, "")
		} else {
			lines = append(lines, "")
		}
		lines = append(lines, s.Group.Render(strings.ToUpper(g.Title)))
		for _, b := range g.Bindings {
			lines = append(lines, formatBinding(b, s))
		}
	}
	lines = append(lines, "", s.Subtle.Render(footerHint))
	return strings.Join(lines, "\n")
}

// formatBinding draws one row: "  <kbd chips padded>  <desc>".
// Keys split on " " so each token gets its own chip; tokens like "or"
// stay as plain glue so "↑ ↓ or k j" reads as four chips with one
// separator word.
func formatBinding(b Binding, s Styles) string {
	keyCell := renderKeys(b.Keys, s)
	pad := max(1, keyColWidth-lipgloss.Width(keyCell))
	return leftIndent + keyCell + strings.Repeat(" ", pad) + s.Desc.Render(b.Desc)
}

// renderKeys turns a "↑ ↓ or k j" specification into chips joined by
// the literal glue tokens (or, /, ,) painted muted.
func renderKeys(keys string, s Styles) string {
	tokens := strings.Fields(keys)
	out := strings.Builder{}
	for i, tok := range tokens {
		if i > 0 {
			out.WriteString(" ")
		}
		if isGlue(tok) {
			out.WriteString(s.Subtle.Render(tok))
			continue
		}
		out.WriteString(ui.Kbd(tok, s.Kbd))
	}
	return out.String()
}

func isGlue(tok string) bool {
	switch tok {
	case "or", "/", ",":
		return true
	}
	return false
}
