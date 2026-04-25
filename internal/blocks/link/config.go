package link

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Config struct {
	Upstream string `json:"upstream"`
	ServeAt  string `json:"serve_at"`
}

var (
	errMissingUpstream = errors.New("link: upstream is required")
	errMissingServeAt  = errors.New("link: serve_at is required")
)

func parseConfig(raw json.RawMessage) (Config, error) {
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("link: parse config: %w", err)
	}
	if c.Upstream == "" {
		return Config{}, errMissingUpstream
	}
	if c.ServeAt == "" {
		return Config{}, errMissingServeAt
	}
	return c, nil
}

func parseInitialControls(raw json.RawMessage) Controls {
	var c Controls
	if len(raw) == 0 {
		return c
	}
	_ = json.Unmarshal(raw, &c)
	return c
}
