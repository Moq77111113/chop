// Package tui hosts the bubbletea-based terminal interface for `chop run`.
package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds the lipgloss styles derived from the chop oklch design tokens.
// One source of truth — every component references styles through Theme.
type Theme struct {
	Bg      lipgloss.Color
	Bg1     lipgloss.Color
	Bg2     lipgloss.Color
	Fg      lipgloss.Color
	Muted   lipgloss.Color
	Dim     lipgloss.Color
	Line    lipgloss.Color
	Line2   lipgloss.Color
	Primary lipgloss.Color
	Warn    lipgloss.Color
	Danger  lipgloss.Color
	Info    lipgloss.Color
	Accent  lipgloss.Color

	Title     lipgloss.Style
	Subtle    lipgloss.Style
	Statusbar lipgloss.Style
	Panel     lipgloss.Style
	Selected  lipgloss.Style
}

// DefaultTheme returns the canonical chop palette as sRGB
func DefaultTheme() Theme {
	t := Theme{
		Bg:      lipgloss.Color("#1c1f29"),
		Bg1:     lipgloss.Color("#23273234"),
		Bg2:     lipgloss.Color("#2a2f3b"),
		Fg:      lipgloss.Color("#e7e9ed"),
		Muted:   lipgloss.Color("#a8acb5"),
		Dim:     lipgloss.Color("#7a7e87"),
		Line:    lipgloss.Color("#3a3e48"),
		Line2:   lipgloss.Color("#4d525c"),
		Primary: lipgloss.Color("#7ad88a"),
		Warn:    lipgloss.Color("#d4b86d"),
		Danger:  lipgloss.Color("#d47770"),
		Info:    lipgloss.Color("#6db7d4"),
		Accent:  lipgloss.Color("#6dd4b7"),
	}
	t.Title = lipgloss.NewStyle().Foreground(t.Fg).Bold(true)
	t.Subtle = lipgloss.NewStyle().Foreground(t.Muted)
	t.Statusbar = lipgloss.NewStyle().Foreground(t.Dim)
	t.Panel = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(t.Line)
	t.Selected = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	return t
}
