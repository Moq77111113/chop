// Package empty renders the left-pane empty state: ASCII silhouette,
// guidance copy, and two example commands users can copy-paste.
package empty

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/ui"
)

// emptyArt is the stylised "no input" silhouette shown when no blocks
// are running. Plain runes — no ANSI — so EmptyArt can colour it as a
// whole.
var emptyArt = []string{
	"    .---.",
	"    |   |",
	".---'---'---.",
	"|  no input  |",
	"'------------'",
}

// emptyExample is a single suggested command rendered as a small
// framed card. The split (cmd / arg) lets the renderer paint the
// user-supplied portion in the primary accent so the call-to-action
// reads at a glance.
type emptyExample struct{ cmd, arg string }

var emptyExamples = []emptyExample{
	{"chop run", "cam.mp4"},
	{"chop run", "rtsp://192.168.1.10:554/live"},
}

const (
	emptyHintCopy = "no source attached. point chop at an rtsp url or a file:"
	emptyAddCopy  = "or press"
	emptyAddTrail = "to add one interactively"
)

// Props is the input bag — empty: the copy is fixed.
type Props struct{}

// Styles bundles the palette: art tint, hint, example frame/segments,
// and the kbd chip used by the "press [a]" hint.
type Styles struct {
	Art            lipgloss.Style
	Hint           lipgloss.Style
	ExampleFrame   lipgloss.Style
	ExamplePrompt  lipgloss.Style
	ExampleCommand lipgloss.Style
	ExampleArg     lipgloss.Style
	Kbd            ui.KbdStyles
}

// Render produces the centered empty-state at width × height.
func Render(_ Props, s Styles, width, height int) string {
	art := s.Art.Render(strings.Join(emptyArt, "\n"))
	hint := s.Hint.Render(emptyHintCopy)
	cards := make([]string, 0, len(emptyExamples))
	for _, ex := range emptyExamples {
		body := s.ExamplePrompt.Render("$") + " " +
			s.ExampleCommand.Render(ex.cmd) + " " +
			s.ExampleArg.Render(ex.arg)
		cards = append(cards, s.ExampleFrame.Render(body))
	}
	addLine := s.Hint.Render(emptyAddCopy) + " " +
		ui.Kbd("a", s.Kbd) + " " +
		s.Hint.Render(emptyAddTrail)
	stack := lipgloss.JoinVertical(lipgloss.Center, art, "", hint, "", cards[0], cards[1], "", addLine)
	if height <= 0 {
		return lipgloss.PlaceHorizontal(width, lipgloss.Center, stack)
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, stack)
}
