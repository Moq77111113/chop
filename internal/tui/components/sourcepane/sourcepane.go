// Package sourcepane renders the right-pane variant when a source
// block is selected: header, "stream source" calm card, RTP-served
// counter, uptime.
package sourcepane

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/data"
)

const (
	cardLabel = "STREAM SOURCE"
	noteMsg   = "no impairments — this block produces the stream."
)

// Props carries the source identity, snapshot, and consumable URL.
type Props struct {
	ID   string
	Type string
	Snap data.SourceSnapshot
	// HasSnap distinguishes "snapshot not yet seen" from "all-zero
	// snapshot is real" so the renderer can show "starting…" for the
	// pre-snapshot tick.
	HasSnap bool
	URL     string
}

// Styles bundles the palette: header, subtle, calm card frame, and
// stat key/value emphasis.
type Styles struct {
	Header  lipgloss.Style
	Subtle  lipgloss.Style
	Frame   lipgloss.Style
	Label   lipgloss.Style
	Note    lipgloss.Style
	StatKey lipgloss.Style
	StatVal lipgloss.Style
}

// Render produces the pane at width × height.
func Render(p Props, s Styles, width, _ int) string {
	header := s.Header.Render(strings.ToUpper(p.ID)) + " " + s.Subtle.Render("("+p.Type+")")
	// Frame has Border(rounded); lipgloss Width(N) excludes border, so
	// to land on `width` cells total we set Width(width-2).
	cardW := max(width-2, 1)
	cardBody := s.Label.Render(cardLabel) + "\n" + s.Note.Render(noteMsg)
	if p.URL != "" {
		cardBody += "\n" + s.Subtle.Render("consume from "+p.URL)
	}
	card := s.Frame.Width(cardW).Render(cardBody)
	return strings.Join([]string{header, "", card, "", renderStats(p, s)}, "\n")
}

func renderStats(p Props, s Styles) string {
	if !p.HasSnap {
		return s.Subtle.Render("starting…")
	}
	uptime := "—"
	if p.Snap.UpSinceMs > 0 {
		uptime = time.Since(time.UnixMilli(p.Snap.UpSinceMs)).Truncate(time.Second).String()
	}
	lines := []string{
		s.StatKey.Render("rtp packets served  ") + s.StatVal.Render(fmt.Sprintf("%d", p.Snap.RTPServed)),
		s.StatKey.Render("up since            ") + s.StatVal.Render(uptime),
	}
	return strings.Join(lines, "\n")
}
