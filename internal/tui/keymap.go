package tui

import "github.com/charmbracelet/bubbles/key"

// Keymap declares every binding the chop TUI handles. The statusbar advertises
// a contextual subset; the full set is the source of truth for the help
// overlay opened with ?.
type Keymap struct {
	Quit       key.Binding
	Help       key.Binding
	Up         key.Binding
	Down       key.Binding
	Decrease   key.Binding
	Increase   key.Binding
	Drill      key.Binding
	Back       key.Binding
	Zero       key.Binding
	ResetLink  key.Binding
	ResetAll   key.Binding
	Copy       key.Binding
	ToggleFeed key.Binding
}

// DefaultKeymap returns the canonical bindings.
func DefaultKeymap() Keymap {
	return Keymap{
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "down")),
		Decrease:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←", "adjust")),
		Increase:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→", "adjust")),
		Drill:      key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("↵", "drill in")),
		Back:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Zero:       key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "zero knob")),
		ResetLink:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reset link")),
		ResetAll:   key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "reset all")),
		Copy:       key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy as flags")),
		ToggleFeed: key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "toggle feed")),
	}
}
