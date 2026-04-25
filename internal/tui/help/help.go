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
	cardWidth   = 64
	keyColWidth = 18
	leftIndent  = "  "
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
	for i, g := range groups {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, st.Group.Render(strings.ToUpper(g.Title)))
		for _, b := range g.Bindings {
			lines = append(lines, formatBinding(b, st))
		}
	}
	lines = append(lines, "", st.Subtle.Render("? or esc to close"))
	return strings.Join(lines, "\n")
}

// formatBinding draws one row: "  <keys padded to keyColWidth>  <desc>".
// Keys are bold-colored inline tokens; no borders, no multi-line cells.
func formatBinding(b Binding, st Styles) string {
	keyCell := st.Key.Render(b.Keys)
	pad := max(1, keyColWidth-lipgloss.Width(keyCell))
	return leftIndent + keyCell + strings.Repeat(" ", pad) + st.Desc.Render(b.Desc)
}
