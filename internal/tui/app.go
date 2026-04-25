package tui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/tui/linklist"
)

const (
	titleSeparator = " · "
	titleSuffix    = "/tʃɒp/"
	tickInterval   = time.Second / 15
	snapshotBudget = 200 * time.Millisecond
)

type tickMsg struct{}

type rowsMsg struct{ rows []linklist.Row }

// App is the root bubbletea model. It owns the layout shell, the focus
// machine, and the polling loop that feeds child components.
type App struct {
	sup      *supervisor.Supervisor
	theme    Theme
	keymap   Keymap
	width    int
	height   int
	linkList *linklist.Model
	listSt   linklist.Styles
}

// New constructs the App bound to a running supervisor. The supervisor is
// expected to be started in a sibling goroutine; the TUI consumes its
// registry on each tick.
func New(sup *supervisor.Supervisor) *App {
	t := DefaultTheme()
	a := &App{
		sup:      sup,
		theme:    t,
		keymap:   DefaultKeymap(),
		linkList: &linklist.Model{},
		listSt:   newListStyles(t),
	}
	return a
}

// Init schedules the first refresh tick.
func (a *App) Init() tea.Cmd { return a.tickCmd() }

// Update handles window resize, polling, and global keys.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keymap.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keymap.PickUp):
			a.linkList.MoveUp()
		case key.Matches(msg, a.keymap.PickDown):
			a.linkList.MoveDown()
		}
	case tickMsg:
		return a, tea.Batch(a.tickCmd(), a.fetchCmd())
	case rowsMsg:
		a.linkList.Set(msg.rows)
	}
	return a, nil
}

// View renders the layout shell: titlebar, body (link list pane today,
// focused-link pane tomorrow), statusbar.
func (a *App) View() string {
	if a.width == 0 {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		a.titlebar(),
		a.body(),
		a.statusbar(),
	)
}

func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// fetchCmd returns a Cmd that snapshots every block in the registry and
// emits a rowsMsg. It runs on bubbletea's command goroutine, never blocking
// the render loop.
func (a *App) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), snapshotBudget)
		defer cancel()
		handles := a.sup.Registry.List()
		rows := make([]linklist.Row, 0, len(handles))
		for _, h := range handles {
			snap, err := h.Snapshot(ctx)
			state := linklist.StateStarting
			if err == nil {
				state = mapState(snap.Status)
			}
			rows = append(rows, linklist.Row{ID: h.ID, Type: h.Type, State: state})
		}
		return rowsMsg{rows: rows}
	}
}

func mapState(s block.Status) linklist.State {
	switch s {
	case block.StatusRunning:
		return linklist.StateUp
	case block.StatusDegraded:
		return linklist.StateDegraded
	case block.StatusStopped:
		return linklist.StateStopped
	}
	return linklist.StateStarting
}

func (a *App) titlebar() string {
	left := a.theme.Title.Render("chop") + a.theme.Subtle.Render(titleSeparator+titleSuffix)
	return lipgloss.NewStyle().
		Width(a.width).
		Foreground(a.theme.Fg).
		Background(a.theme.Bg1).
		Padding(0, 1).
		Render(left)
}

func (a *App) body() string {
	bodyHeight := max(a.height-2, 1)
	return lipgloss.NewStyle().
		Width(a.width).
		Height(bodyHeight).
		Padding(1, 2).
		Render(a.linkList.Render(a.listSt, a.width-4, bodyHeight-2))
}

func (a *App) statusbar() string {
	hints := []string{
		a.keymap.PickUp.Help().Key + " pick link",
		a.keymap.NextKnob.Help().Key + " pick knob",
		a.keymap.Increase.Help().Key + " adjust",
		a.keymap.Quit.Help().Key + " quit",
	}
	return a.theme.Statusbar.
		Width(a.width).
		Background(a.theme.Bg1).
		Padding(0, 1).
		Render(strings.Join(hints, titleSeparator))
}

func newListStyles(t Theme) linklist.Styles {
	return linklist.Styles{
		Header:   lipgloss.NewStyle().Foreground(t.Muted).Bold(true).MarginBottom(1),
		Type:     lipgloss.NewStyle().Foreground(t.Dim),
		Selected: lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Empty:    lipgloss.NewStyle().Foreground(t.Dim).Italic(true),
		StateUp:  lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		StateDeg: lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		StateBad: lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
	}
}
