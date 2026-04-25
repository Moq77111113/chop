package tui

import (
	"encoding/json"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/scenario"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/tui/focused"
	"github.com/moq77111113/chop/internal/tui/help"
	"github.com/moq77111113/chop/internal/tui/linklist"
)

const (
	titleSeparator = " · "
	titleSuffix    = "/tʃɒp/"
	bootMessage    = "starting chop · spawning blocks…"
	connectingHint = "spawning blocks · waiting for first snapshot…"
)

type uiState struct {
	focusOn      focusZone
	helpOpen     bool
	confirmReset bool
	coachOpen    bool
	toastMsg     string
	toastUntil   time.Time
	firstReady   bool
}

type dataState struct {
	links   map[string]linkSnapshot
	sources map[string]sourceSnapshot
	history map[string][]float64
	events  []uiEvent
	linksAt time.Time
}

type uiStyles struct {
	list   linklist.Styles
	focus  focused.Styles
	help   help.Styles
	events eventStyles
}

// App is the root bubbletea model. Sibling files: keys.go (input), poll.go
// (data flow), layout.go (rendering), styles.go (theme bindings).
type App struct {
	sup    *supervisor.Supervisor
	theme  Theme
	keymap Keymap

	width, height int
	list          *linklist.Model
	focused       *focused.Model

	configs map[string]json.RawMessage

	ui     uiState
	data   dataState
	styles uiStyles
}

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
		list:    &linklist.Model{},
		focused: focused.New(),
		configs: configs,
		styles: uiStyles{
			list:   newListStyles(t),
			focus:  newFocusedStyles(t),
			help:   newHelpStyles(t),
			events: newEventStyles(t),
		},
		data: dataState{
			links:   map[string]linkSnapshot{},
			sources: map[string]sourceSnapshot{},
			history: map[string][]float64{},
		},
		ui: uiState{coachOpen: shouldShowCoach()},
	}
}

func (a *App) Init() tea.Cmd { return tea.Batch(a.fetchCmd(), a.tickCmd()) }

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
		if !a.ui.toastUntil.IsZero() && !time.Now().Before(a.ui.toastUntil) {
			a.ui.toastMsg = ""
			a.ui.toastUntil = time.Time{}
		}
	}
	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		return a.theme.Subtle.Render(bootMessage)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		a.titlebar(),
		a.body(),
		a.statusbar(),
	)
}

// consumableURL returns the rtsp endpoint a downstream consumer (ffplay,
// mediamtx, etc.) should hit for a given block. Empty string when the
// block isn't known or doesn't expose one.
func (a *App) consumableURL(id string) string {
	raw, ok := a.configs[id]
	if !ok {
		return ""
	}
	var lc struct {
		ServeAt string `json:"serve_at"`
	}
	if json.Unmarshal(raw, &lc) == nil && lc.ServeAt != "" {
		return "rtsp://" + lc.ServeAt
	}
	var sc struct {
		Listen string `json:"listen"`
	}
	if json.Unmarshal(raw, &sc) == nil && sc.Listen != "" {
		return "rtsp://" + sc.Listen
	}
	return ""
}
