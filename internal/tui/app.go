package tui

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/scenario"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/tui/components/coach"
	"github.com/moq77111113/chop/internal/tui/components/empty"
	"github.com/moq77111113/chop/internal/tui/components/events"
	"github.com/moq77111113/chop/internal/tui/components/focused"
	"github.com/moq77111113/chop/internal/tui/components/help"
	"github.com/moq77111113/chop/internal/tui/components/linklist"
	"github.com/moq77111113/chop/internal/tui/components/sourcepane"
	"github.com/moq77111113/chop/internal/tui/components/statusbar"
	"github.com/moq77111113/chop/internal/tui/components/titlebar"
	"github.com/moq77111113/chop/internal/tui/components/toast"
	"github.com/moq77111113/chop/internal/tui/data"
	"github.com/moq77111113/chop/internal/tui/knobs"
	"github.com/moq77111113/chop/internal/tui/state"
)

const (
	bootMessage    = "starting chop · spawning blocks…"
	connectingHint = "spawning blocks · waiting for first snapshot…"
	devVersion     = "dev"
)

// uiStyles bundles the per-component palettes. Each constructor lives
// in styles.go.
type uiStyles struct {
	titlebar   titlebar.Styles
	statusbar  statusbar.Styles
	emptyPane  empty.Styles
	sourcePane sourcepane.Styles
	list       linklist.Styles
	focus      focused.Styles
	help       help.Styles
	coach      coach.Styles
	toast      toast.Styles
	events     events.Styles
}

// App is the root bubbletea model. Sibling files: keys.go (input),
// poll.go (data flow), layout.go (rendering), styles.go (theme bindings).
type App struct {
	sup    *supervisor.Supervisor
	theme  Theme
	keymap Keymap

	width, height int
	rows          []data.Row
	cursor        state.Cursor
	pane          *knobs.Pane

	configs map[string]json.RawMessage

	ui      state.UI
	links   map[string]data.LinkSnapshot
	sources map[string]data.SourceSnapshot
	history *state.History
	events  []uiEvent
	linksAt time.Time

	styles uiStyles
}

// New builds an App. It binds the supervisor for the data flow, the
// scenario for config lookups, and seeds the default theme/keymap.
func New(sup *supervisor.Supervisor, sc *scenario.Scenario) *App {
	t := DefaultTheme()
	configs := map[string]json.RawMessage{}
	if sc != nil {
		for _, b := range sc.Blocks {
			configs[b.ID] = b.Config
		}
	}
	return &App{
		sup:     sup,
		theme:   t,
		keymap:  DefaultKeymap(),
		pane:    knobs.NewPane(),
		configs: configs,
		links:   map[string]data.LinkSnapshot{},
		sources: map[string]data.SourceSnapshot{},
		history: state.NewHistory(),
		styles: uiStyles{
			titlebar:   newTitlebarStyles(t),
			statusbar:  newStatusbarStyles(t),
			emptyPane:  newEmptyStyles(t),
			sourcePane: newSourcepaneStyles(t),
			list:       newLinklistStyles(t),
			focus:      newFocusedStyles(t),
			help:       newHelpStyles(t),
			coach:      newCoachStyles(t),
			toast:      newToastStyles(t),
			events:     newEventsStyles(t),
		},
		ui: state.UI{CoachOpen: shouldShowCoach()},
	}
}

// Init starts the supervisor poll cadence and the time tick.
func (a *App) Init() tea.Cmd { return tea.Batch(a.fetchCmd(), a.tickCmd()) }

// Update is the bubbletea reducer.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case tea.KeyMsg:
		return a, a.handleKey(msg)
	case tickMsg:
		return a, tea.Batch(a.tickCmd(), a.fetchCmd())
	case rowsMsg:
		a.applyRowsMsg(msg)
	case EventMsg:
		a.appendEvent(msg.Event)
	case toastClearMsg:
		a.ui.MaybeClearToast(time.Now())
	}
	return a, nil
}

// View renders the full TUI frame.
func (a *App) View() string {
	if a.width == 0 {
		return a.theme.Subtle.Render(bootMessage)
	}
	parts := []string{a.titlebar(), a.body()}
	if a.ui.Toast != "" {
		parts = append(parts, a.toastBanner())
	}
	parts = append(parts, a.statusbar())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// titlebar composes the top chrome row from titlebar.Render. The
// summary, URL, time, and version come from app state and helpers; the
// component itself stays stateless.
func (a *App) titlebar() string {
	url := ""
	if r, ok := a.selectedRow(); ok {
		url = a.consumableURL(r.ID)
	}
	t := time.Now().Format("15:04:05")
	if a.isNarrow() {
		t = ""
	}
	return titlebar.Render(titlebar.Props{
		Summary:       a.titlebarSummary(),
		ConsumableURL: url,
		Time:          t,
		Version:       version(),
		Narrow:        a.isNarrow(),
	}, a.styles.titlebar, a.width)
}

// titlebarSummary collapses the registry into a one-line readout: link
// count + global state ("3 links · running" or "3 links · 1 down").
func (a *App) titlebarSummary() string {
	if len(a.rows) == 0 {
		return "idle"
	}
	down := 0
	for _, r := range a.rows {
		if r.State == data.StateDown || r.State == data.StateStopped {
			down++
		}
	}
	if down > 0 {
		return fmt.Sprintf("%d links · %d down", len(a.rows), down)
	}
	return fmt.Sprintf("%d links · running", len(a.rows))
}

// statusbar composes the bottom chrome row. The hints / body source
// lives in keys.go's statusbarProps so the keymap and the render path
// stay co-located.
func (a *App) statusbar() string {
	return statusbar.Render(a.statusbarProps(), a.styles.statusbar, a.width)
}

// version reads the build info for the right-cluster stamp on the
// titlebar. Falls back to "dev" when the binary was built without
// module / vcs info.
func version() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return devVersion
	}
	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return devVersion
}

// consumableURL is a thin lookup that defers the actual config-parse
// to data.ConsumableURL. Empty string when the block isn't known.
func (a *App) consumableURL(id string) string {
	raw, ok := a.configs[id]
	if !ok {
		return ""
	}
	return data.ConsumableURL(raw)
}

// selectedRow returns the row under the cursor, or ok=false when no
// row is selected (empty registry).
func (a *App) selectedRow() (data.Row, bool) {
	idx := a.cursor.Selected(len(a.rows))
	if idx < 0 {
		return data.Row{}, false
	}
	return a.rows[idx], true
}
