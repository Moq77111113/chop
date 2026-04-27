package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/data"
)

// PillStyles holds the per-state foreground styles for the link state
// badge. Up renders "UP" in the calm/primary tint, Degraded warns,
// Bad alarms.
type PillStyles struct {
	Up, Degraded, Bad lipgloss.Style
}

// Pill renders the state badge: UP / DEGRADED / DOWN / STARTING /
// STOPPED in the appropriate foreground style.
func Pill(stateValue data.LinkState, s PillStyles) string {
	switch stateValue {
	case data.StateUp:
		return s.Up.Render(string(stateValue))
	case data.StateDegraded:
		return s.Degraded.Render(string(stateValue))
	case data.StateDown, data.StateStopped:
		return s.Bad.Render(string(stateValue))
	case data.StateStarting:
		return s.Degraded.Render(string(stateValue))
	}
	return string(stateValue)
}
