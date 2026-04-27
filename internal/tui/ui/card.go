package ui

import "github.com/charmbracelet/lipgloss"

// CardVariant selects the border tint of a card.
type CardVariant int

// Card variants — neutral / cautious / alarming / muted.
const (
	CardCalm CardVariant = iota
	CardWarn
	CardBad
	CardDim
)

// CardStyles holds the four border palettes indexed by variant.
type CardStyles struct {
	Frame [4]lipgloss.Style
}

// Card renders content inside a rounded border. Width 0 means "size to
// content".
func Card(content string, variant CardVariant, s CardStyles, width int) string {
	style := s.Frame[variant]
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(content)
}
