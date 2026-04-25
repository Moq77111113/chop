package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/focused"
	"github.com/moq77111113/chop/internal/tui/help"
	"github.com/moq77111113/chop/internal/tui/intent"
	"github.com/moq77111113/chop/internal/tui/knob"
	"github.com/moq77111113/chop/internal/tui/linklist"
)

// newListStyles builds the linklist palette from the chop theme tokens.
func newListStyles(t Theme) linklist.Styles {
	rowBase := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginBottom(1)
	return linklist.Styles{
		Header:           lipgloss.NewStyle().Foreground(t.Muted).Bold(true).MarginBottom(1),
		Type:             lipgloss.NewStyle().Foreground(t.Dim),
		Title:            lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Selected:         lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Summary:          lipgloss.NewStyle().Foreground(t.Dim),
		Spark:            lipgloss.NewStyle().Foreground(t.Primary),
		Empty:            lipgloss.NewStyle().Foreground(t.Dim).Italic(true),
		StateUp:          lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		StateDeg:         lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		StateBad:         lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		RowFrame:         rowBase.BorderForeground(t.Line),
		RowFrameSelected: rowBase.BorderForeground(t.Primary),
		EmptyArt:         lipgloss.NewStyle().Foreground(t.Dim),
		EmptyText:        lipgloss.NewStyle().Foreground(t.Muted),
		ExampleFrame:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(0, 1),
		ExamplePrompt:    lipgloss.NewStyle().Foreground(t.Dim),
		ExampleCommand:   lipgloss.NewStyle().Foreground(t.Fg),
		ExampleArg:       lipgloss.NewStyle().Foreground(t.Primary),
	}
}

// newHelpStyles builds the help-overlay palette. Plain inline keys: bold
// primary-colored tokens, no borders — bordered keys broke vertical layout.
func newHelpStyles(t Theme) help.Styles {
	return help.Styles{
		Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(1, 2),
		Title:  lipgloss.NewStyle().Foreground(t.Fg).Bold(true).MarginBottom(1),
		Group:  lipgloss.NewStyle().Foreground(t.Muted).Bold(true),
		Key:    lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Desc:   lipgloss.NewStyle().Foreground(t.Fg),
		Subtle: lipgloss.NewStyle().Foreground(t.Dim).Italic(true),
	}
}

// newFocusedStyles builds the right-pane palette: link header, knob track,
// intent strip, and reset bar.
func newFocusedStyles(t Theme) focused.Styles {
	return focused.Styles{
		Header:     lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Subtle:     lipgloss.NewStyle().Foreground(t.Dim),
		Divider:    lipgloss.NewStyle().Foreground(t.Line),
		Knob:       newKnobStyles(t),
		Intent:     newIntentStyles(t),
		ResetFrame: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(0, 1),
		ResetLabel: lipgloss.NewStyle().Foreground(t.Muted),
		ResetKey:   lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
	}
}

func newKnobStyles(t Theme) knob.Styles {
	return knob.Styles{
		Label:   lipgloss.NewStyle().Foreground(t.Muted).Bold(true),
		LabelOn: lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Hint:    lipgloss.NewStyle().Foreground(t.Primary),
		Track:   lipgloss.NewStyle().Foreground(t.Line),
		Fill: [3]lipgloss.Style{
			lipgloss.NewStyle().Foreground(t.Primary),
			lipgloss.NewStyle().Foreground(t.Warn),
			lipgloss.NewStyle().Foreground(t.Danger),
		},
		Value: [3]lipgloss.Style{
			lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
			lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
			lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		},
		Scale:    lipgloss.NewStyle().Foreground(t.Dim),
		FocusBar: lipgloss.NewStyle().Foreground(t.Primary),
	}
}

func newIntentStyles(t Theme) intent.Styles {
	frame := func(c lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c).Padding(0, 1)
	}
	return intent.Styles{
		Calm:    lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Warning: lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		Bad:     lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		Body:    lipgloss.NewStyle().Foreground(t.Fg),
		Number:  lipgloss.NewStyle().Foreground(t.Primary),
		Frame: [3]lipgloss.Style{
			frame(t.Line2),
			frame(t.Warn),
			frame(t.Danger),
		},
	}
}
