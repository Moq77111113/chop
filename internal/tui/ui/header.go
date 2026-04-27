package ui

import "github.com/charmbracelet/lipgloss"

// Header renders an uppercase, letter-spaced section caption. Kept as
// a primitive so callers don't need to know the recipe — if the design
// evolves (icon prefix, baseline rule), it absorbs here once.
func Header(text string, s lipgloss.Style) string { return s.Render(text) }
