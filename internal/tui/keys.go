package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/moq77111113/chop/internal/tui/components/help"
	"github.com/moq77111113/chop/internal/tui/components/statusbar"
	"github.com/moq77111113/chop/internal/tui/state"
)

func (a *App) handleKey(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, a.keymap.Quit) {
		return tea.Quit
	}
	if a.ui.CoachOpen {
		a.ui.DismissCoach()
		go markSeen()
		// fall through — the dismissing keypress also performs its action
	}
	if a.ui.ConfirmReset {
		return a.handleConfirmKey(msg)
	}
	if a.ui.HelpOpen {
		if key.Matches(msg, a.keymap.Back) || key.Matches(msg, a.keymap.Help) {
			a.ui.CloseHelp()
		}
		return nil
	}
	if key.Matches(msg, a.keymap.Help) {
		a.ui.OpenHelp()
		return nil
	}
	if key.Matches(msg, a.keymap.ResetAll) {
		a.ui.BeginConfirmReset()
		return nil
	}
	if a.ui.Focus == state.FocusList {
		return a.handleListKey(msg)
	}
	return a.handleKnobsKey(msg)
}

func (a *App) handleConfirmKey(msg tea.KeyMsg) tea.Cmd {
	a.ui.ResolveConfirmReset()
	if msg.String() == "y" {
		return a.resetAllCmd()
	}
	return nil
}

func (a *App) handleListKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, a.keymap.Drill):
		a.ui.Focus = state.FocusKnobs
	case key.Matches(msg, a.keymap.Up):
		a.cursor.MoveUp()
		a.syncFocusedFromSelection()
	case key.Matches(msg, a.keymap.Down):
		a.cursor.MoveDown(len(a.rows))
		a.syncFocusedFromSelection()
	}
	return nil
}

func (a *App) handleKnobsKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, a.keymap.Back), key.Matches(msg, a.keymap.Drill):
		a.ui.Focus = state.FocusList
	case key.Matches(msg, a.keymap.Up):
		a.pane.PrevKnob()
	case key.Matches(msg, a.keymap.Down):
		a.pane.NextKnob()
	case key.Matches(msg, a.keymap.Increase):
		return a.applyCmd(a.pane.Adjust(+1))
	case key.Matches(msg, a.keymap.Decrease):
		return a.applyCmd(a.pane.Adjust(-1))
	case key.Matches(msg, a.keymap.Zero):
		return a.applyCmd(a.pane.Zero())
	case key.Matches(msg, a.keymap.ResetLink):
		return a.applyCmd(a.pane.ResetAll())
	case key.Matches(msg, a.keymap.Copy):
		return a.copyAsFlagsCmd()
	}
	return nil
}

// statusbarProps builds the bottom-bar payload: a free-form Body for
// transient prompts (confirm-reset), or a contextual list of (key,
// label) hint chips otherwise.
func (a *App) statusbarProps() statusbar.Props {
	if a.ui.ConfirmReset {
		count := len(a.rows)
		return statusbar.Props{Body: fmt.Sprintf("reset %d block(s)? [y/n]", count)}
	}
	if a.ui.Focus == state.FocusKnobs {
		return statusbar.Props{Hints: []statusbar.Hint{
			{Key: "↑↓", Label: "pick knob"},
			{Key: "←→", Label: "adjust"},
			{Key: "y", Label: "copy"},
			{Key: "r", Label: "reset"},
			{Key: "esc", Label: "back"},
			{Key: "?", Label: "help"},
		}}
	}
	return statusbar.Props{Hints: []statusbar.Hint{
		{Key: "↑↓", Label: "pick link"},
		{Key: "↵", Label: "drill in"},
		{Key: "?", Label: "help"},
		{Key: "q", Label: "quit"},
	}}
}

func (a *App) helpGroups() []help.Group {
	return []help.Group{
		{Title: "navigate", Bindings: []help.Binding{
			{Keys: "↑ ↓ / k j", Desc: "move within the active pane"},
			{Keys: "↵ / tab", Desc: "drill into knobs / toggle pane"},
			{Keys: "esc", Desc: "back to the link list"},
		}},
		{Title: "adjust", Bindings: []help.Binding{
			{Keys: "← → / h l", Desc: "decrease / increase the focused knob"},
			{Keys: "0", Desc: "zero the focused knob"},
			{Keys: "r", Desc: "reset all knobs on the focused link"},
			{Keys: "R", Desc: "reset every link"},
		}},
		{Title: "capture", Bindings: []help.Binding{
			{Keys: "y", Desc: "copy current perturbation as flags"},
		}},
		{Title: "app", Bindings: []help.Binding{
			{Keys: "?", Desc: "toggle this help"},
			{Keys: "q / ctrl+c", Desc: "quit"},
		}},
	}
}
