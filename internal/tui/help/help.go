// Package help renders the keymap card opened with ? from the chop TUI.
// The card replaces the body content when visible; dismissing returns to the
// previous layout.
package help

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// Styles bundles the lipgloss styles the help card draws with.
type Styles struct {
	Border lipgloss.Style
	Title  lipgloss.Style
	Group  lipgloss.Style
	Key    lipgloss.Style
	Desc   lipgloss.Style
	Subtle lipgloss.Style
}

const (
	cardWidth      = 60
	keyCellWidth   = 22
	keyTokenSep    = " "
)

// Render draws the help card at width × height. If the available area is
// smaller than the card, the card is clipped from the right and bottom.
func Render(groups []Group, st Styles, width, height int) string {
	body := bodyLines(groups, st)
	card := st.Border.Width(min(cardWidth, width-2)).Render(body)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
}

func bodyLines(groups []Group, st Styles) string {
	var lines []string
	lines = append(lines, st.Title.Render("chop · keymap"))
	lines = append(lines, "")
	for _, g := range groups {
		lines = append(lines, st.Group.Render(strings.ToUpper(g.Title)))
		for _, b := range g.Bindings {
			lines = append(lines, formatBinding(b, st))
		}
		lines = append(lines, "")
	}
	lines = append(lines, st.Subtle.Render("? or esc to close"))
	return strings.Join(lines, "\n")
}

func formatBinding(b Binding, st Styles) string {
	keyCell := renderKeyCells(b.Keys, st)
	pad := strings.Repeat(" ", max(2, keyCellWidth-lipgloss.Width(keyCell)))
	return "  " + keyCell + pad + st.Desc.Render(b.Desc)
}

// renderKeyCells splits a Keys string on whitespace, then renders the
// modifier-free tokens as small bordered boxes and leaves separators
// like "/" or "or" as plain text — matching the design's keys-as-keys
// look.
func renderKeyCells(keys string, st Styles) string {
	tokens := strings.Fields(keys)
	parts := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if isSeparator(tok) {
			parts = append(parts, st.Desc.Render(tok))
			continue
		}
		parts = append(parts, st.Key.Render(tok))
	}
	return strings.Join(parts, keyTokenSep)
}

func isSeparator(tok string) bool { return tok == "/" || tok == "or" }
