package source

import (
	"encoding/json"
	"errors"
	"fmt"
)

const defaultFPS = 25

type Config struct {
	File   string `json:"file"`
	Listen string `json:"listen"`
	FPS    int    `json:"fps"`
}

var (
	errMissingFile   = errors.New("source: file is required")
	errMissingListen = errors.New("source: listen is required")
)

func parseConfig(raw json.RawMessage) (Config, error) {
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("source: parse config: %w", err)
	}
	if c.File == "" {
		return Config{}, errMissingFile
	}
	if c.Listen == "" {
		return Config{}, errMissingListen
	}
	if c.FPS == 0 {
		c.FPS = defaultFPS
	}
	return c, nil
}
