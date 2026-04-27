package data

import "encoding/json"

// linkConsumable matches a link block's `serve_at` field — the host:port
// the proxy advertises downstream consumers should pull from.
type linkConsumable struct {
	ServeAt string `json:"serve_at"`
}

// sourceConsumable matches a source block's `listen` field — the
// host:port the source's RTSP server is bound to.
type sourceConsumable struct {
	Listen string `json:"listen"`
}

// ConsumableURL returns the rtsp endpoint a downstream consumer (ffplay,
// mediamtx, etc.) should hit for the given block config. Empty string
// when the config doesn't expose either field.
func ConsumableURL(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var lc linkConsumable
	if json.Unmarshal(raw, &lc) == nil && lc.ServeAt != "" {
		return "rtsp://" + lc.ServeAt
	}
	var sc sourceConsumable
	if json.Unmarshal(raw, &sc) == nil && sc.Listen != "" {
		return "rtsp://" + sc.Listen
	}
	return ""
}
