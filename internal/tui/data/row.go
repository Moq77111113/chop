package data

// Block-type discriminator used by titles, link list components, and
// renderer dispatch. Stay strings — they match the Type field on
// supervisor.Handle (the wire identity).
const (
	BlockTypeLink    = "link"
	BlockTypeSource  = "source"
	BlockTypeProcess = "process"
)

// Row is the cross-layer view of one block-list entry. Logic packages
// (data, state) read and decorate it; components consume it directly
// in their Props. No fields above what every renderer actually needs.
type Row struct {
	ID    string
	Type  string
	Title string // "src → dst" or "" → fall back to ID
	State LinkState
	Knobs Values
	Spark []float64
	Rate  string // "2.4 Mb/s" or "— Mb/s"
}
