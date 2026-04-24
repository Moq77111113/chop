// Package scenario parses and validates chop YAML scenarios.
// Pure — no I/O beyond reading the YAML bytes. Safe to call from `chop lint`.
package scenario

import "encoding/json"

type Scenario struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Blocks      []Block `yaml:"blocks"`
}

type Block struct {
	ID       string          `yaml:"id"`
	Type     string          `yaml:"type"`
	Binary   string          `yaml:"binary,omitempty"` // for external block types
	Config   json.RawMessage `yaml:"config,omitempty"`
	Controls json.RawMessage `yaml:"controls,omitempty"`
}
