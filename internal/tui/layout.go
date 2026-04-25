package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/help"
	"github.com/moq77111113/chop/internal/tui/linklist"
)

const (
	rightPaneMin    = 38 // narrow → stack
	rightPaneMax    = 48 // wide  → fixed width
	leftMinShare    = 0.55
	narrowThreshold = 80 // ≤ → stack panes vertically

	sourceNoteMsg = "no impairments — this block produces the stream."
)

func (a *App) titlebar() string {
	left := a.titlebarLeft()
	right := a.titlebarRight()
	gap := max(a.width-lipgloss.Width(left)-lipgloss.Width(right)-2, 1)
	bar := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().
		Width(a.width).
		Foreground(a.theme.Fg).
		Background(a.theme.Bg1).
		Padding(0, 1).
		Render(bar)
}

func (a *App) titlebarLeft() string {
	dot := lipgloss.NewStyle().Foreground(a.theme.Primary).Render("●")
	name := a.theme.Title.Render("chop")
	ipa := a.theme.Subtle.Render("/tʃɒp/")
	return strings.Join([]string{
		dot + " " + name,
		ipa,
		a.theme.Subtle.Render(a.titlebarSummary()),
	}, titleSeparator)
}

func (a *App) titlebarSummary() string {
	rows := a.list.Rows()
	if len(rows) == 0 {
		return "idle"
	}
	down := 0
	for _, r := range rows {
		if r.State == linklist.StateDown || r.State == linklist.StateStopped {
			down++
		}
	}
	if down > 0 {
		return fmt.Sprintf("%d links · %d down", len(rows), down)
	}
	return fmt.Sprintf("%d links · running", len(rows))
}

func (a *App) titlebarRight() string {
	rows := a.list.Rows()
	parts := []string{time.Now().Format("15:04:05")}
	if len(rows) > 0 {
		idx := max(0, a.list.Selected())
		if url := a.consumableURL(rows[idx].ID); url != "" {
			parts = append([]string{url}, parts...)
		}
	}
	return a.theme.Subtle.Render(strings.Join(parts, titleSeparator))
}

func (a *App) body() string {
	bodyHeight := max(a.height-2, 1)
	if a.ui.coachOpen {
		return renderCoach(a.width, bodyHeight, a.theme)
	}
	if a.ui.helpOpen {
		return help.Render(a.helpGroups(), a.styles.help, a.width, bodyHeight)
	}
	if !a.ui.firstReady {
		return lipgloss.NewStyle().
			Width(a.width).
			Height(bodyHeight).
			Padding(1, 2).
			Render(a.theme.Subtle.Render(connectingHint))
	}
	if a.width <= narrowThreshold {
		return a.bodyStacked(bodyHeight)
	}
	return a.bodySideBySide(bodyHeight)
}

func (a *App) bodySideBySide(bodyHeight int) string {
	rightW := rightPaneWidth(a.width)
	leftW := a.width - rightW
	innerH := max(bodyHeight-2, 1)

	left := a.paneFrame(leftW, bodyHeight, a.ui.focusOn == focusList).Render(
		a.leftPaneContent(leftW-4, innerH),
	)
	right := a.paneFrame(rightW, bodyHeight, a.ui.focusOn == focusKnobs).Render(
		a.rightPaneContent(rightW-4, innerH),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// bodyStacked renders the panes top-to-bottom for narrow terminals. The
// link list takes ~40% of the height; the focused pane gets the rest. The
// focus rail still hugs the left edge of whichever pane has focus.
func (a *App) bodyStacked(bodyHeight int) string {
	topH := max(bodyHeight*2/5, 6)
	if topH > bodyHeight-4 {
		topH = max(bodyHeight-4, 1)
	}
	botH := max(bodyHeight-topH, 1)
	innerW := max(a.width-4, 1)

	top := a.paneFrame(a.width, topH, a.ui.focusOn == focusList).Render(
		a.leftPaneContent(innerW, max(topH-2, 1)),
	)
	bot := a.paneFrame(a.width, botH, a.ui.focusOn == focusKnobs).Render(
		a.rightPaneContent(innerW, max(botH-2, 1)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, bot)
}

// paneFrame wraps each pane. The focused pane gets a primary 1-cell left
// rail; the other gets a transparent placeholder so widths stay aligned.
func (a *App) paneFrame(width, height int, focused bool) lipgloss.Style {
	rail := a.theme.Bg
	if focused {
		rail = a.theme.Primary
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1, 1, 2).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(rail)
}

func (a *App) rightPaneContent(width, height int) string {
	rows := a.list.Rows()
	idx := a.list.Selected()
	if idx < 0 || idx >= len(rows) {
		return a.theme.Subtle.Render("no block selected")
	}
	r := rows[idx]
	url := a.consumableURL(r.ID)
	switch r.Type {
	case blockTypeLink:
		body := a.focused.Render(a.styles.focus, width, height, a.intentSnapshot())
		if url != "" {
			body = a.theme.Subtle.Render("consume from "+url) + "\n\n" + body
		}
		return body
	case blockTypeSource:
		return a.renderSourcePane(r, url)
	}
	return a.theme.Subtle.Render("unknown block type: " + r.Type)
}

func (a *App) renderSourcePane(r linklist.Row, url string) string {
	header := a.styles.focus.Header.Render(strings.ToUpper(r.ID)) + " " + a.styles.focus.Subtle.Render("("+r.Type+")")
	note := lipgloss.NewStyle().Foreground(a.theme.Muted).Italic(true).Render(sourceNoteMsg)
	parts := []string{header, ""}
	if url != "" {
		parts = append(parts, a.theme.Subtle.Render("consume from "+url), "")
	}
	parts = append(parts, note, "", a.sourceStats(r.ID))
	return strings.Join(parts, "\n")
}

func (a *App) sourceStats(id string) string {
	s, ok := a.data.sources[id]
	if !ok {
		return a.theme.Subtle.Render("starting…")
	}
	uptime := "—"
	if s.UpSinceMs > 0 {
		uptime = time.Since(time.UnixMilli(s.UpSinceMs)).Truncate(time.Second).String()
	}
	keyStyle := lipgloss.NewStyle().Foreground(a.theme.Muted)
	valStyle := lipgloss.NewStyle().Foreground(a.theme.Fg).Bold(true)
	lines := []string{
		keyStyle.Render("rtp packets served  ") + valStyle.Render(fmt.Sprintf("%d", s.RTPServed)),
		keyStyle.Render("up since            ") + valStyle.Render(uptime),
	}
	return strings.Join(lines, "\n")
}

// leftPaneContent stacks the link list on top of the events ticker.
func (a *App) leftPaneContent(width, height int) string {
	feed := renderEventsTicker(a.data.events, a.styles.events, width)
	if feed == "" {
		return a.list.Render(a.styles.list, width, height)
	}
	feedH := lipgloss.Height(feed)
	listH := max(height-feedH-1, 1)
	return a.list.Render(a.styles.list, width, listH) + "\n" + feed
}

func rightPaneWidth(total int) int {
	share := total - int(float64(total)*leftMinShare)
	if share < rightPaneMin {
		return rightPaneMin
	}
	if share > rightPaneMax {
		return rightPaneMax
	}
	return share
}

func (a *App) statusbar() string {
	if a.ui.toastMsg != "" {
		return lipgloss.NewStyle().
			Width(a.width).
			Foreground(a.theme.Bg).
			Background(a.theme.Primary).
			Bold(true).
			Padding(0, 1).
			Render(a.ui.toastMsg)
	}
	hints := a.statusHints()
	return a.theme.Statusbar.
		Width(a.width).
		Background(a.theme.Bg1).
		Padding(0, 1).
		Render(strings.Join(hints, titleSeparator))
}
