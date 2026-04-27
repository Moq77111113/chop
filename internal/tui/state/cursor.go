package state

// Cursor is a length-aware selection index. Length is supplied by the
// caller on each call so the cursor never needs to track row count
// separately.
type Cursor struct{ idx int }

// Selected returns the cursor index, or -1 when length is 0.
func (c Cursor) Selected(length int) int {
	if length == 0 {
		return -1
	}
	return c.idx
}

// MoveUp clamps at 0.
func (c *Cursor) MoveUp() {
	if c.idx > 0 {
		c.idx--
	}
}

// MoveDown clamps at length-1.
func (c *Cursor) MoveDown(length int) {
	if c.idx < length-1 {
		c.idx++
	}
}

// Clamp keeps the index inside [0, length-1] across registry changes.
func (c *Cursor) Clamp(length int) {
	if c.idx >= length {
		c.idx = length - 1
	}
	if c.idx < 0 {
		c.idx = 0
	}
}
