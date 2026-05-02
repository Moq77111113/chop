package scenario

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const blockTypeProcess = "process"

// Load reads, parses, and validates a scenario file. Returns *Scenario or a structured error.
func Load(path string) (*Scenario, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s Scenario
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := validate(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func validate(s *Scenario) error {
	if len(s.Blocks) == 0 {
		return fmt.Errorf("scenario has no blocks")
	}
	seen := make(map[string]bool)
	for i, b := range s.Blocks {
		if b.ID == "" {
			return fmt.Errorf("block #%d: id required", i)
		}
		if b.Type == "" {
			return fmt.Errorf("block %s: type required", b.ID)
		}
		if seen[b.ID] {
			return fmt.Errorf("duplicate block id: %s", b.ID)
		}
		seen[b.ID] = true
		if err := validateBlockConfig(b); err != nil {
			return err
		}
	}
	return nil
}

func validateBlockConfig(b Block) error {
	if b.Type != blockTypeProcess {
		return nil
	}
	var cfg struct {
		Cmd string `json:"cmd"`
	}
	if len(b.Config) > 0 {
		if err := json.Unmarshal(b.Config, &cfg); err != nil {
			return fmt.Errorf("block %s: parse config: %w", b.ID, err)
		}
	}
	if cfg.Cmd == "" {
		return fmt.Errorf("block %s: cmd is required for process blocks", b.ID)
	}
	return nil
}
