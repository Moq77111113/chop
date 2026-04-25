package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

const toastDuration = 2500 * time.Millisecond

type toastClearMsg struct{}

// copyAsFlagsCmd serialises the focused link's controls as a `--override`
// flag string and writes it to the system clipboard. A toast confirms the
// copy (or surfaces a clipboard error if xclip / wl-copy / equivalent is
// missing).
func (a *App) copyAsFlagsCmd() tea.Cmd {
	id := a.focused.LinkID
	if id == "" {
		return a.showToastCmd("nothing to copy — pick a link first")
	}
	snap, ok := a.data.links[id]
	if !ok {
		return a.showToastCmd("waiting on first snapshot…")
	}
	flag := formatOverrideFlag(id, snap)
	if err := clipboard.WriteAll(flag); err != nil {
		return a.showToastCmd("clipboard unavailable: " + err.Error())
	}
	return a.showToastCmd("copied · " + flag)
}

func (a *App) showToastCmd(msg string) tea.Cmd {
	a.ui.toastMsg = msg
	a.ui.toastUntil = time.Now().Add(toastDuration)
	return tea.Tick(toastDuration, func(time.Time) tea.Msg { return toastClearMsg{} })
}

// formatOverrideFlag emits the canonical `<id>:<knob>=<val>,…` form. All
// four knobs are always written so a paste-back round-trips deterministically
// regardless of the originating block's defaults.
func formatOverrideFlag(id string, snap linkSnapshot) string {
	const (
		percentScale = 100.0
		kbpsToMbps   = 1000.0
	)
	parts := []string{
		fmt.Sprintf("loss=%.0f%%", snap.Controls.Loss*percentScale),
		fmt.Sprintf("latency=%dms", snap.Controls.LatencyMs),
		fmt.Sprintf("jitter=%dms", snap.Controls.JitterMs),
		"bw=" + bandwidthFlag(snap.Controls.BandwidthKbps, kbpsToMbps),
	}
	return id + ":" + strings.Join(parts, ",")
}

func bandwidthFlag(kbps uint32, perMbps float64) string {
	if kbps == 0 {
		return "off"
	}
	return fmt.Sprintf("%.1fM", float64(kbps)/perMbps)
}
