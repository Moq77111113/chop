package linklist

import "github.com/charmbracelet/lipgloss"

// Styles is the small surface the linklist needs from the parent theme.
// We pass styles in (rather than importing the tui Theme) so the package
// stays leaf-level and easy to test in isolation.
type Styles struct {
	Header           lipgloss.Style
	Type             lipgloss.Style
	Title            lipgloss.Style
	Selected         lipgloss.Style
	Summary          lipgloss.Style
	Spark            lipgloss.Style
	Empty            lipgloss.Style
	StateUp          lipgloss.Style
	StateDeg         lipgloss.Style
	StateBad         lipgloss.Style
	RowFrame         lipgloss.Style
	RowFrameSelected lipgloss.Style
	// Empty-state palette. The left pane shows an ASCII illustration plus
	// two example-command cards when no blocks are running yet.
	EmptyArt       lipgloss.Style
	EmptyText      lipgloss.Style
	ExampleFrame   lipgloss.Style
	ExamplePrompt  lipgloss.Style
	ExampleCommand lipgloss.Style
	ExampleArg     lipgloss.Style
}

// statePill picks the right style for a state value.
func (s Styles) statePill(st State) string {
	switch st {
	case StateUp:
		return s.StateUp.Render(string(StateUp))
	case StateDegraded:
		return s.StateDeg.Render(string(StateDegraded))
	case StateDown, StateStopped:
		return s.StateBad.Render(string(st))
	case StateStarting:
		return s.StateDeg.Render(string(StateStarting))
	}
	return string(st)
}
