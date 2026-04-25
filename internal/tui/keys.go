package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/moq77111113/chop/internal/tui/help"
)

type focusZone int

const (
	focusList focusZone = iota
	focusKnobs
)

func (a *App) handleKey(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, a.keymap.Quit) {
		return tea.Quit
	}
	if a.ui.coachOpen {
		a.ui.coachOpen = false
		go markSeen()
		// fall through — the dismissing keypress also performs its action
	}
	if a.ui.confirmReset {
		return a.handleConfirmKey(msg)
	}
	if a.ui.helpOpen {
		if key.Matches(msg, a.keymap.Back) || key.Matches(msg, a.keymap.Help) {
			a.ui.helpOpen = false
		}
		return nil
	}
	if key.Matches(msg, a.keymap.Help) {
		a.ui.helpOpen = true
		return nil
	}
	if key.Matches(msg, a.keymap.ResetAll) {
		a.ui.confirmReset = true
		return nil
	}
	if a.ui.focusOn == focusList {
		return a.handleListKey(msg)
	}
	return a.handleKnobsKey(msg)
}

func (a *App) handleConfirmKey(msg tea.KeyMsg) tea.Cmd {
	a.ui.confirmReset = false
	if msg.String() == "y" {
		return a.resetAllCmd()
	}
	return nil
}

func (a *App) handleListKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, a.keymap.Drill):
		a.ui.focusOn = focusKnobs
	case key.Matches(msg, a.keymap.Up):
		a.list.MoveUp()
		a.syncFocusedFromList()
	case key.Matches(msg, a.keymap.Down):
		a.list.MoveDown()
		a.syncFocusedFromList()
	}
	return nil
}

func (a *App) handleKnobsKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, a.keymap.Back), key.Matches(msg, a.keymap.Drill):
		a.ui.focusOn = focusList
	case key.Matches(msg, a.keymap.Up):
		a.focused.PrevKnob()
	case key.Matches(msg, a.keymap.Down):
		a.focused.NextKnob()
	case key.Matches(msg, a.keymap.Increase):
		return a.applyCmd(a.focused.Adjust(+1))
	case key.Matches(msg, a.keymap.Decrease):
		return a.applyCmd(a.focused.Adjust(-1))
	case key.Matches(msg, a.keymap.Zero):
		return a.applyCmd(a.focused.Zero())
	case key.Matches(msg, a.keymap.ResetLink):
		return a.applyCmd(a.focused.ResetAll())
	case key.Matches(msg, a.keymap.Copy):
		return a.copyAsFlagsCmd()
	}
	return nil
}

func (a *App) statusHints() []string {
	if a.ui.confirmReset {
		count := len(a.list.Rows())
		return []string{fmt.Sprintf("reset %d block(s)? [y/n]", count)}
	}
	if a.ui.focusOn == focusKnobs {
		return []string{
			"↑↓ pick knob",
			"←→ adjust",
			"y copy",
			"r reset",
			"esc back",
			"? help",
		}
	}
	return []string{
		"↑↓ pick link",
		"↵ drill in",
		"? help",
		"q quit",
	}
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
