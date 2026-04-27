package ui

import "github.com/charmbracelet/lipgloss"

// KbdStyles bundles the chip palette. Frame paints the surrounding
// chrome (vertical bars, border, or inverse background — whichever the
// theme picks); Key paints the inner text.
type KbdStyles struct {
	Frame lipgloss.Style
	Key   lipgloss.Style
}

// Kbd renders a keyboard hint as a chip showing the key text.
//
// Example: ui.Kbd("↑↓", styles) → "│↑↓│" with vertical-bar sides.
func Kbd(keyText string, s KbdStyles) string {
	return s.Frame.Render(s.Key.Render(keyText))
}
