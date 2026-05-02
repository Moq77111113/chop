package tui

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/components/coach"
	"github.com/moq77111113/chop/internal/tui/components/empty"
	"github.com/moq77111113/chop/internal/tui/components/events"
	"github.com/moq77111113/chop/internal/tui/components/focused"
	"github.com/moq77111113/chop/internal/tui/components/help"
	"github.com/moq77111113/chop/internal/tui/components/linklist"
	"github.com/moq77111113/chop/internal/tui/components/processpane"
	"github.com/moq77111113/chop/internal/tui/components/sourcepane"
	"github.com/moq77111113/chop/internal/tui/components/toast"
	"github.com/moq77111113/chop/internal/tui/data"
	"github.com/moq77111113/chop/internal/tui/state"
)

const (
	rightPaneMin    = 38 // narrow → stack
	rightPaneMax    = 48 // wide  → fixed width
	leftMinShare    = 0.55
	narrowThreshold = 80 // ≤ → stack panes vertically
)

// isNarrow returns true when the terminal is too tight to stack panes
// horizontally; the same threshold drives both the body stacking and
// the titlebar's metadata trim so the two stay in sync.
func (a *App) isNarrow() bool { return a.width <= narrowThreshold }

func (a *App) body() string {
	chrome := 2
	if a.ui.Toast != "" {
		chrome++
	}
	bodyHeight := max(a.height-chrome, 1)
	if !a.ui.FirstReady {
		return lipgloss.NewStyle().
			Width(a.width).
			Height(bodyHeight).
			Padding(1, 2).
			Render(a.theme.Subtle.Render(connectingHint))
	}

	base := a.bodyContent(bodyHeight)
	if a.ui.HelpOpen {
		modal := help.Render(help.Props{Groups: a.helpGroups(), Version: version()},
			a.styles.help, a.width, bodyHeight)
		return overlayCenter(dimBehind(base), modal)
	}
	if a.ui.CoachOpen {
		modal := coach.Render(coach.Props{}, a.styles.coach, a.width, bodyHeight)
		return overlayCenter(dimBehind(base), modal)
	}
	return base
}

// bodyContent renders just the link list / focused pane. The overlay
// wrappers (help, coach) stack on top in body().
func (a *App) bodyContent(bodyHeight int) string {
	if a.width <= narrowThreshold {
		return a.bodyStacked(bodyHeight)
	}
	return a.bodySideBySide(bodyHeight)
}

func (a *App) bodySideBySide(bodyHeight int) string {
	rightW := rightPaneWidth(a.width)
	leftW := a.width - rightW
	innerH := max(bodyHeight-2, 1)

	left := a.paneFrame(leftW, bodyHeight, a.ui.Focus == state.FocusList).Render(
		a.leftPaneContent(leftW-paneInnerOverhead, innerH),
	)
	right := a.paneFrame(rightW, bodyHeight, a.ui.Focus == state.FocusKnobs).Render(
		a.rightPaneContent(rightW-paneInnerOverhead, innerH),
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
	innerW := max(a.width-paneInnerOverhead, 1)

	top := a.paneFrame(a.width, topH, a.ui.Focus == state.FocusList).Render(
		a.leftPaneContent(innerW, max(topH-2, 1)),
	)
	bot := a.paneFrame(a.width, botH, a.ui.Focus == state.FocusKnobs).Render(
		a.rightPaneContent(innerW, max(botH-2, 1)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, bot)
}

// paneFrame wraps each pane. The focused pane gets a primary 1-cell
// left rail; the other gets a transparent placeholder so widths stay
// aligned. Padding is symmetric (top/bottom 1, left/right 2) so the
// content has the right-edge breathing room shown in
// design/screens.html.
func (a *App) paneFrame(width, height int, focusRail bool) lipgloss.Style {
	rail := a.theme.Bg
	if focusRail {
		rail = a.theme.Primary
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2, 1, 2).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(rail)
}

// paneInnerOverhead is the visual budget the pane chrome eats from the
// outer width: 1 border (left) + 2 padding (left) + 2 padding (right).
const paneInnerOverhead = 5

func (a *App) rightPaneContent(width, height int) string {
	r, ok := a.selectedRow()
	if !ok {
		return a.renderEmptySelection(width, height)
	}
	url := a.consumableURL(r.ID)
	switch r.Type {
	case data.BlockTypeLink:
		body := focused.Render(focused.Props{
			LinkID:   r.ID,
			LinkType: r.Type,
			Pane:     a.pane,
			Intent:   a.intentSnapshot(),
		}, a.styles.focus, width, height)
		if url != "" {
			body = a.theme.Subtle.Render("consume from "+url) + "\n\n" + body
		}
		return body
	case data.BlockTypeSource:
		snap, has := a.sources[r.ID]
		return sourcepane.Render(sourcepane.Props{
			ID:      r.ID,
			Type:    r.Type,
			Snap:    snap,
			HasSnap: has,
			URL:     url,
		}, a.styles.sourcePane, width, height)
	case data.BlockTypeProcess:
		snap, has := a.processes[r.ID]
		return processpane.Render(processpane.Props{
			ID:      r.ID,
			Type:    r.Type,
			Snap:    snap,
			HasSnap: has,
			Cmd:     a.processCmdLine(r.ID),
			Cwd:     a.processCwd(r.ID),
		}, a.styles.processPane, width, height)
	}
	return a.theme.Subtle.Render("unknown block type: " + r.Type)
}

func (a *App) processCmdLine(id string) string {
	raw, ok := a.configs[id]
	if !ok {
		return ""
	}
	var pc struct {
		Cmd  string   `json:"cmd"`
		Args []string `json:"args"`
	}
	if json.Unmarshal(raw, &pc) != nil {
		return ""
	}
	if len(pc.Args) == 0 {
		return pc.Cmd
	}
	return pc.Cmd + " " + strings.Join(pc.Args, " ")
}

func (a *App) processCwd(id string) string {
	raw, ok := a.configs[id]
	if !ok {
		return ""
	}
	var pc struct {
		Cwd string `json:"cwd"`
	}
	if json.Unmarshal(raw, &pc) != nil {
		return ""
	}
	return pc.Cwd
}

// leftPaneContent stacks the link list on top of the events ticker.
// Narrow mode drops the ticker — there's not enough room for the four
// columns to read, and the focused pane below already carries the
// per-link narrative. When no rows are present we hand off to the
// empty-pane component.
func (a *App) leftPaneContent(width, height int) string {
	if len(a.rows) == 0 {
		return empty.Render(empty.Props{}, a.styles.emptyPane, width, height)
	}
	listProps := linklist.Props{Rows: a.rows, Selected: a.cursor.Selected(len(a.rows))}
	if a.isNarrow() {
		return linklist.Render(listProps, a.styles.list, width, height)
	}
	feed := events.Render(toEventList(a.events), a.styles.events, width)
	if feed == "" {
		return linklist.Render(listProps, a.styles.list, width, height)
	}
	feedH := lipgloss.Height(feed)
	listH := max(height-feedH-1, 1)
	return linklist.Render(listProps, a.styles.list, width, listH) + "\n" + feed
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

// toastBanner delegates to the toast component. Returns "" when no
// message is set, in which case the View won't include this row.
func (a *App) toastBanner() string {
	return toast.Render(toast.Props{
		Message: a.ui.Toast,
		Flag:    a.ui.ToastFlag,
	}, a.styles.toast, a.width)
}
