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
	parts := []string{dot + " " + name, ipa}
	if !a.isNarrow() {
		parts = append(parts, a.theme.Subtle.Render(a.titlebarSummary()))
	}
	return strings.Join(parts, titleSeparator)
}

// isNarrow returns true when the terminal is too tight to stack panes
// horizontally; the same threshold drives both the body stacking and
// the titlebar's metadata trim so the two stay in sync.
func (a *App) isNarrow() bool { return a.width <= narrowThreshold }

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
	if a.isNarrow() {
		return ""
	}
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
	chrome := 2
	if a.ui.toastMsg != "" {
		chrome++
	}
	bodyHeight := max(a.height-chrome, 1)
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
		return a.renderEmptySelection(width, height)
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
		return a.renderSourcePane(r, url, width)
	}
	return a.theme.Subtle.Render("unknown block type: " + r.Type)
}

const sourceCardLabel = "STREAM SOURCE"

// renderSourcePane mirrors the link pane's vertical rhythm — header on
// top, calm-bordered card with the explanatory copy, then live stats —
// so a source selection doesn't read as a different screen.
func (a *App) renderSourcePane(r linklist.Row, url string, width int) string {
	header := a.styles.focus.Header.Render(strings.ToUpper(r.ID)) + " " + a.styles.focus.Subtle.Render("("+r.Type+")")
	cardW := max(width, 1)
	cardBody := lipgloss.NewStyle().Foreground(a.theme.Primary).Bold(true).Render(sourceCardLabel) + "\n" +
		lipgloss.NewStyle().Foreground(a.theme.Fg).Render(sourceNoteMsg)
	if url != "" {
		cardBody += "\n" + a.theme.Subtle.Render("consume from "+url)
	}
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.theme.Line2).
		Padding(0, 1).
		Width(cardW).
		Render(cardBody)
	return strings.Join([]string{header, "", card, "", a.sourceStats(r.ID)}, "\n")
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
// Narrow mode drops the ticker — there's not enough room for the four
// columns to read, and the focused pane below already carries the
// per-link narrative.
func (a *App) leftPaneContent(width, height int) string {
	if a.isNarrow() {
		return a.list.Render(a.styles.list, width, height)
	}
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
	hints := a.statusHints()
	return a.theme.Statusbar.
		Width(a.width).
		Background(a.theme.Bg1).
		Padding(0, 1).
		Render(strings.Join(hints, titleSeparator))
}

// renderEmptySelection is the right-pane counterpart to the link list's
// empty card: a "SELECTED · —" header above a hint that nothing is wired
// up yet. Mirrors the wording in design/screens/01-empty.png.
func (a *App) renderEmptySelection(width, height int) string {
	header := a.styles.focus.Header.Render("SELECTED") + " " + a.styles.focus.Subtle.Render("· —")
	body := a.theme.Subtle.Render("no link selected.") + "\n\n" +
		a.theme.Subtle.Render("controls appear here once a source attaches.")
	stack := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	if height <= 0 {
		return stack
	}
	return lipgloss.Place(width, height, lipgloss.Right, lipgloss.Center, stack)
}

// toastBanner is the transient row rendered above the statusbar. Two
// flavours: a simple centered pill for one-line messages (errors, hints),
// and the richer copy-as-flags card — chip + truncated CLI + paste hint.
func (a *App) toastBanner() string {
	if a.ui.toastFlag != "" {
		return a.copyToastBanner()
	}
	badge := lipgloss.NewStyle().
		Foreground(a.theme.Bg).
		Background(a.theme.Primary).
		Bold(true).
		Padding(0, 1).
		Render("✓ " + a.ui.toastMsg)
	return lipgloss.PlaceHorizontal(a.width, lipgloss.Center, badge,
		lipgloss.WithWhitespaceBackground(a.theme.Bg1))
}

const (
	copyToastPasteHint = "· paste anywhere"
	copyToastEllipsis  = "…"
	copyToastChromeW   = 8 // chip + gaps + dim hint padding budget
)

// copyToastBanner renders the rich copy-as-flags card centered above the
// statusbar: green "✓ copied" chip on the left, the would-be CLI prefixed
// with `chop run --override …` in the middle (truncated with ellipsis to
// fit width), and a dim "· paste anywhere" hint on the right.
func (a *App) copyToastBanner() string {
	chip := lipgloss.NewStyle().
		Foreground(a.theme.Bg).
		Background(a.theme.Primary).
		Bold(true).
		Padding(0, 1).
		Render("✓ " + a.ui.toastMsg)
	hint := lipgloss.NewStyle().Foreground(a.theme.Dim).Italic(true).Render(copyToastPasteHint)
	cmdW := max(a.width-lipgloss.Width(chip)-lipgloss.Width(hint)-copyToastChromeW, 8)
	cmd := truncateMiddle(a.ui.toastFlag, cmdW)
	cmdBox := lipgloss.NewStyle().
		Foreground(a.theme.Fg).
		Background(a.theme.Bg2).
		Padding(0, 1).
		Render(cmd)
	row := lipgloss.JoinHorizontal(lipgloss.Center, chip, " ", cmdBox, " ", hint)
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.theme.Primary).
		Padding(0, 1).
		Render(row)
	return lipgloss.PlaceHorizontal(a.width, lipgloss.Center, card,
		lipgloss.WithWhitespaceBackground(a.theme.Bg1))
}

// truncateMiddle clips an ANSI-free string to width with a trailing
// ellipsis. Used by the copy toast — the override flag's prefix carries
// the most identifying info (block id, loss), so we keep the head and
// drop the tail rather than the other way around.
func truncateMiddle(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= lipgloss.Width(copyToastEllipsis) {
		return copyToastEllipsis
	}
	keep := width - lipgloss.Width(copyToastEllipsis)
	return s[:keep] + copyToastEllipsis
}
