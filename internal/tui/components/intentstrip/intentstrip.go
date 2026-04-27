// Package intentstrip renders the "stream looks like" strip under the
// focused link header. Variants: calm / warn / bad. Reads from the
// pure-logic intent.Snapshot via intent.Classify.
package intentstrip

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/intent"
)

// Styles bundles the per-variant palette.
type Styles struct {
	// Frame is indexed by intent severity (Calm / Warn / Bad).
	Frame [3]lipgloss.Style
	// Label colors indexed the same way.
	Label [3]lipgloss.Style
	// Body is the foreground of the description copy.
	Body lipgloss.Style
}

// Render frames the classifier's verdict in the appropriate variant.
// Frame has Border(rounded) + Padding(0,1); lipgloss Width(N) excludes
// border but includes padding, so we set Width(width-2) to land on a
// total of width cells. Body wraps at width-4 to leave the padding
// untouched.
func Render(s intent.Snapshot, st Styles, width int) string {
	v := intent.Classify(s)
	innerW := max(width-4, 1)
	body := wrap(st.Body.Render(v.Body), innerW)
	content := st.Label[v.Severity].Render(v.Label) + "\n" + body
	frame := st.Frame[v.Severity]
	return frame.Width(max(width-2, 1)).Render(content)
}

func wrap(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for _, w := range words {
		candidate := cur
		if cur != "" {
			candidate += " "
		}
		candidate += w
		if lipgloss.Width(candidate) > width && cur != "" {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur = candidate
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}
