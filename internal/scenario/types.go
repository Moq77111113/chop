// Package scenario parses and validates chop YAML scenarios.
// Pure — no I/O beyond reading the YAML bytes. Safe to call from `chop lint`.
package scenario

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// Scenario is the parsed and validated YAML scenario file.
type Scenario struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Blocks      []Block `yaml:"blocks"`
}

// Block is one declared block in a scenario. Config and Controls are kept as
// opaque json.RawMessage so they can be passed through to the block child
// without the scenario package having to know each block type's schema.
type Block struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Binary   string          `json:"binary,omitempty"`
	Config   json.RawMessage `json:"config,omitempty"`
	Controls json.RawMessage `json:"controls,omitempty"`
}

// UnmarshalYAML decodes the YAML block, then re-marshals the nested
// Config/Controls maps as JSON so downstream consumers see a single shape
// (json.RawMessage) regardless of source format.
func (b *Block) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		ID       string `yaml:"id"`
		Type     string `yaml:"type"`
		Binary   string `yaml:"binary,omitempty"`
		Config   any    `yaml:"config,omitempty"`
		Controls any    `yaml:"controls,omitempty"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	cfg, err := encodeJSON(raw.Config)
	if err != nil {
		return err
	}
	ctrl, err := encodeJSON(raw.Controls)
	if err != nil {
		return err
	}
	b.ID = raw.ID
	b.Type = raw.Type
	b.Binary = raw.Binary
	b.Config = cfg
	b.Controls = ctrl
	return nil
}

func encodeJSON(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
