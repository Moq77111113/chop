package tui

import "github.com/charmbracelet/bubbles/key"

// Keymap declares every binding the chop TUI handles. The statusbar advertises
// a contextual subset; the full set is the source of truth for the help
// overlay (M2.5).
type Keymap struct {
	Quit       key.Binding
	Help       key.Binding
	PickUp     key.Binding
	PickDown   key.Binding
	NextKnob   key.Binding
	PrevKnob   key.Binding
	Decrease   key.Binding
	Increase   key.Binding
	Zero       key.Binding
	ResetLink  key.Binding
	ResetAll   key.Binding
	Copy       key.Binding
	ToggleFeed key.Binding
}

// DefaultKeymap returns the canonical bindings spec'd in DESIGN.md §5.
func DefaultKeymap() Keymap {
	return Keymap{
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		PickUp:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑", "pick link")),
		PickDown:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "pick link")),
		NextKnob:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "pick knob")),
		PrevKnob:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev knob")),
		Decrease:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←", "adjust")),
		Increase:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→", "adjust")),
		Zero:       key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "zero knob")),
		ResetLink:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reset link")),
		ResetAll:   key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "reset all")),
		Copy:       key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy as flags")),
		ToggleFeed: key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "toggle feed")),
	}
}
