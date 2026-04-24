package block

import (
	"context"
	"encoding/json"
	"testing"
)

// noopBlock implements Block with the minimum.
type noopBlock struct{}

func (noopBlock) Info() Info                               { return Info{} }
func (noopBlock) Snapshot() Snapshot                       { return Snapshot{} }
func (noopBlock) Apply(_ json.RawMessage) error            { return nil }
func (noopBlock) Action(_ string, _ json.RawMessage) error { return nil }
func (noopBlock) Run(ctx context.Context) error            { <-ctx.Done(); return nil }

// TestBlock_InterfaceIsImplementable: si la signature change et casse,
// ce test ne compile plus → blocker immédiat sur la surface publique.
func TestBlock_InterfaceIsImplementable(t *testing.T) {
	var _ Block = noopBlock{}
}
