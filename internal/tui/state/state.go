// Package state is the TUI's session memory: which modal is open,
// where the focus lives, when the toast expires, where the cursor is,
// and the rolling rate samples driving sparklines. It owns no
// rendering and imports no lipgloss or bubbletea types.
package state

import "time"

// FocusZone selects which pane absorbs movement keys.
type FocusZone int

// FocusZone constants — the two horizontal zones the user navigates
// between with enter/tab/esc.
const (
	FocusList FocusZone = iota
	FocusKnobs
)

// UI captures every piece of UI session state. Methods are tiny
// transitions; the bubbletea adapter calls them, never mutates fields
// directly.
type UI struct {
	Focus        FocusZone
	HelpOpen     bool
	CoachOpen    bool
	ConfirmReset bool
	Toast        string
	ToastFlag    string
	ToastUntil   time.Time
	FirstReady   bool
}

// OpenHelp marks the help modal as visible.
func (u *UI) OpenHelp() { u.HelpOpen = true }

// CloseHelp dismisses the help modal.
func (u *UI) CloseHelp() { u.HelpOpen = false }

// DismissCoach hides the first-run guidance overlay.
func (u *UI) DismissCoach() { u.CoachOpen = false }

// BeginConfirmReset puts the UI in "y/n confirm" mode for reset-all.
func (u *UI) BeginConfirmReset() { u.ConfirmReset = true }

// ResolveConfirmReset clears the confirm-reset prompt.
func (u *UI) ResolveConfirmReset() { u.ConfirmReset = false }

// ShowToast posts a simple chip toast and arms its expiry.
func (u *UI) ShowToast(msg string, until time.Time) {
	u.Toast = msg
	u.ToastFlag = ""
	u.ToastUntil = until
}

// ShowCopyToast posts the rich copy-as-flags toast variant. The chip
// reads "copied"; flag carries the would-be CLI to display.
func (u *UI) ShowCopyToast(flag string, until time.Time) {
	u.Toast = "copied"
	u.ToastFlag = flag
	u.ToastUntil = until
}

// MaybeClearToast clears the toast iff its expiry has passed.
func (u *UI) MaybeClearToast(now time.Time) {
	if !u.ToastUntil.IsZero() && !now.Before(u.ToastUntil) {
		u.Toast = ""
		u.ToastFlag = ""
		u.ToastUntil = time.Time{}
	}
}
