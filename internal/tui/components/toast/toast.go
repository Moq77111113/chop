// Package toast renders the bottom transient banner: a simple chip for
// short messages, or the rich copy-as-flags card showing the would-be
// CLI and a "paste anywhere" hint.
package toast

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	pasteHint = "· paste anywhere"
	ellipsis  = "…"
	chromeW   = 8 // chip + gaps + dim hint padding budget
)

// Props selects the variant via Flag — empty string → simple chip,
// non-empty → copy-as-flags card with the flag string in the middle.
type Props struct {
	Message string
	Flag    string
}

// Styles bundles both variants' palette.
type Styles struct {
	BarBg     lipgloss.Color // surrounding row background colour
	Chip      lipgloss.Style
	CardFrame lipgloss.Style
	CmdBox    lipgloss.Style
	PasteHint lipgloss.Style
}

// Render returns the banner centered at width cells, or "" when no
// message is set.
func Render(p Props, s Styles, width int) string {
	if p.Message == "" {
		return ""
	}
	if p.Flag == "" {
		return renderSimple(p, s, width)
	}
	return renderCopy(p, s, width)
}

func renderSimple(p Props, s Styles, width int) string {
	chip := s.Chip.Render("✓ " + p.Message)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, chip,
		lipgloss.WithWhitespaceBackground(s.BarBg))
}

func renderCopy(p Props, s Styles, width int) string {
	chip := s.Chip.Render("✓ " + p.Message)
	hint := s.PasteHint.Render(pasteHint)
	cmdW := max(width-lipgloss.Width(chip)-lipgloss.Width(hint)-chromeW, 8)
	cmdBox := s.CmdBox.Render(truncateMiddle(p.Flag, cmdW))
	row := lipgloss.JoinHorizontal(lipgloss.Center, chip, " ", cmdBox, " ", hint)
	card := s.CardFrame.Render(row)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, card,
		lipgloss.WithWhitespaceBackground(s.BarBg))
}

// truncateMiddle clips an ANSI-free string to width with a trailing
// ellipsis. The override flag's prefix carries the most identifying
// info (block id, loss), so we keep the head and drop the tail.
func truncateMiddle(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= lipgloss.Width(ellipsis) {
		return ellipsis
	}
	keep := width - lipgloss.Width(ellipsis)
	return s[:keep] + ellipsis
}
