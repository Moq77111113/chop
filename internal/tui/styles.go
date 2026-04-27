package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/components/coach"
	"github.com/moq77111113/chop/internal/tui/components/empty"
	"github.com/moq77111113/chop/internal/tui/components/events"
	"github.com/moq77111113/chop/internal/tui/components/focused"
	"github.com/moq77111113/chop/internal/tui/components/help"
	"github.com/moq77111113/chop/internal/tui/components/intentstrip"
	"github.com/moq77111113/chop/internal/tui/components/knob"
	"github.com/moq77111113/chop/internal/tui/components/linkcard"
	"github.com/moq77111113/chop/internal/tui/components/linklist"
	"github.com/moq77111113/chop/internal/tui/components/sourcepane"
	"github.com/moq77111113/chop/internal/tui/components/statusbar"
	"github.com/moq77111113/chop/internal/tui/components/titlebar"
	"github.com/moq77111113/chop/internal/tui/components/toast"
	"github.com/moq77111113/chop/internal/tui/ui"
)

func newKbdStyles(t Theme) ui.KbdStyles {
	return ui.KbdStyles{
		Frame: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, true).
			BorderForeground(t.Line2),
		Key: lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
	}
}

// newSparkStyles builds the sparkline palette (single foreground
// shared across all rows).
func newSparkStyles(t Theme) ui.SparkStyles {
	return ui.SparkStyles{Bar: lipgloss.NewStyle().Foreground(t.Primary)}
}

// newPillStyles builds the link-state pill palette.
func newPillStyles(t Theme) ui.PillStyles {
	return ui.PillStyles{
		Up:       lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Degraded: lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		Bad:      lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
	}
}

// newTitlebarStyles builds the top-bar palette.
func newTitlebarStyles(t Theme) titlebar.Styles {
	return titlebar.Styles{
		Bar:    lipgloss.NewStyle().Foreground(t.Fg).Background(t.Bg1),
		Dot:    lipgloss.NewStyle().Foreground(t.Primary),
		Name:   lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		IPA:    lipgloss.NewStyle().Foreground(t.Dim),
		Subtle: lipgloss.NewStyle().Foreground(t.Muted),
	}
}

// newStatusbarStyles builds the bottom-bar palette with kbd chips.
func newStatusbarStyles(t Theme) statusbar.Styles {
	return statusbar.Styles{
		Bar:   lipgloss.NewStyle().Foreground(t.Muted).Background(t.Bg1),
		Body:  lipgloss.NewStyle().Foreground(t.Muted),
		Right: lipgloss.NewStyle().Foreground(t.Dim),
		Kbd:   newKbdStyles(t),
	}
}

// newEmptyStyles builds the left-pane empty-state palette.
func newEmptyStyles(t Theme) empty.Styles {
	return empty.Styles{
		Art:            lipgloss.NewStyle().Foreground(t.Dim),
		Hint:           lipgloss.NewStyle().Foreground(t.Muted),
		ExampleFrame:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(0, 1),
		ExamplePrompt:  lipgloss.NewStyle().Foreground(t.Dim),
		ExampleCommand: lipgloss.NewStyle().Foreground(t.Fg),
		ExampleArg:     lipgloss.NewStyle().Foreground(t.Primary),
		Kbd:            newKbdStyles(t),
	}
}

// newSourcepaneStyles builds the right-pane source-block palette.
func newSourcepaneStyles(t Theme) sourcepane.Styles {
	return sourcepane.Styles{
		Header:  lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Subtle:  lipgloss.NewStyle().Foreground(t.Dim),
		Frame:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(0, 1),
		Label:   lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Note:    lipgloss.NewStyle().Foreground(t.Fg),
		StatKey: lipgloss.NewStyle().Foreground(t.Muted),
		StatVal: lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
	}
}

func newLinkcardStyles(t Theme) linkcard.Styles {
	return linkcard.Styles{
		Row: lipgloss.NewStyle().Padding(0, 2),
		RowSelected: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 1),
		Title:       lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Cursor:      lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Rate:        lipgloss.NewStyle().Foreground(t.Dim),
		KnobsLabel:  lipgloss.NewStyle().Foreground(t.Muted),
		KnobsOK:     lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		KnobsDim:    lipgloss.NewStyle().Foreground(t.Dim),
		KnobsWarn:   lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		KnobsDanger: lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		Spark:       newSparkStyles(t),
		Pill:        newPillStyles(t),
	}
}

// newLinklistStyles builds the list palette: header + the full
// linkcard palette nested in.
func newLinklistStyles(t Theme) linklist.Styles {
	return linklist.Styles{
		Header: lipgloss.NewStyle().Foreground(t.Muted).Bold(true).MarginBottom(1),
		Card:   newLinkcardStyles(t),
	}
}

func newHelpStyles(t Theme) help.Styles {
	return help.Styles{
		Frame:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(1, 2),
		Title:  lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		IPA:    lipgloss.NewStyle().Foreground(t.Dim).MarginBottom(1),
		Group:  lipgloss.NewStyle().Foreground(t.Muted).Bold(true),
		Desc:   lipgloss.NewStyle().Foreground(t.Fg),
		Subtle: lipgloss.NewStyle().Foreground(t.Dim).Italic(true),
		Kbd:    newKbdStyles(t),
	}
}

// newCoachStyles builds the first-run overlay palette.
func newCoachStyles(t Theme) coach.Styles {
	return coach.Styles{
		Chip:  lipgloss.NewStyle().Foreground(t.Primary).Background(t.Bg).Bold(true).Padding(0, 1),
		Title: lipgloss.NewStyle().Foreground(t.Bg).Background(t.Primary).Bold(true),
		Cue:   lipgloss.NewStyle().Foreground(t.Bg).Background(t.Primary).Italic(true),
		Frame: lipgloss.NewStyle().Background(t.Primary).Padding(0, 1),
	}
}

// newToastStyles builds the bottom transient banner palette.
func newToastStyles(t Theme) toast.Styles {
	return toast.Styles{
		BarBg: t.Bg1,
		Chip: lipgloss.NewStyle().
			Foreground(t.Bg).
			Background(t.Primary).
			Bold(true).
			Padding(0, 1),
		CardFrame: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 1),
		CmdBox: lipgloss.NewStyle().
			Foreground(t.Fg).
			Background(t.Bg2).
			Padding(0, 1),
		PasteHint: lipgloss.NewStyle().Foreground(t.Dim).Italic(true),
	}
}

// newKnobStyles builds the components/knob palette.
func newKnobStyles(t Theme) knob.Styles {
	return knob.Styles{
		Label:   lipgloss.NewStyle().Foreground(t.Muted).Bold(true),
		LabelOn: lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Hint:    lipgloss.NewStyle().Foreground(t.Primary),
		Value: [3]lipgloss.Style{
			lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
			lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
			lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		},
		Scale: lipgloss.NewStyle().Foreground(t.Dim),
		Track: ui.TrackStyles{
			Fill: [3]lipgloss.Style{
				lipgloss.NewStyle().Foreground(t.Primary),
				lipgloss.NewStyle().Foreground(t.Warn),
				lipgloss.NewStyle().Foreground(t.Danger),
			},
			Empty:    lipgloss.NewStyle().Foreground(t.Line),
			FocusBar: lipgloss.NewStyle().Foreground(t.Primary),
		},
	}
}

// newIntentstripStyles builds the intent strip palette.
func newIntentstripStyles(t Theme) intentstrip.Styles {
	frame := func(c lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c).Padding(0, 1)
	}
	return intentstrip.Styles{
		Frame: [3]lipgloss.Style{
			frame(t.Line2),
			frame(t.Warn),
			frame(t.Danger),
		},
		Label: [3]lipgloss.Style{
			lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
			lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
			lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		},
		Body: lipgloss.NewStyle().Foreground(t.Fg),
	}
}

// newFocusedStyles builds the right-pane palette.
func newFocusedStyles(t Theme) focused.Styles {
	return focused.Styles{
		Header: lipgloss.NewStyle().Foreground(t.Fg).Bold(true),
		Subtle: lipgloss.NewStyle().Foreground(t.Dim),
		Knob:   newKnobStyles(t),
		Intent: newIntentstripStyles(t),
		Reset: focused.ResetStyles{
			Frame: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Line2).Padding(0, 1),
			Label: lipgloss.NewStyle().Foreground(t.Muted),
			Kbd:   newKbdStyles(t),
		},
	}
}

// newEventsStyles builds the events ticker palette.
func newEventsStyles(t Theme) events.Styles {
	return events.Styles{
		Header: lipgloss.NewStyle().Foreground(t.Muted).Bold(true).MarginTop(1),
		Time:   lipgloss.NewStyle().Foreground(t.Dim),
		Source: lipgloss.NewStyle().Foreground(t.Muted),
		Body:   lipgloss.NewStyle().Foreground(t.Fg),
		Inf:    lipgloss.NewStyle().Foreground(t.Info).Bold(true),
		Wrn:    lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		Err:    lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
	}
}
