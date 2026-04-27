package data

import (
	"encoding/json"
	"net/url"
)

// LinkTitles maps each link block id to a "<source-id> → <link-id>"
// label by matching its upstream URL host:port against the listen
// host:port of source blocks. Links with no matching source produce no
// entry; the renderer falls back to the link id.
func LinkTitles(rows []Row, configs map[string]json.RawMessage) map[string]string {
	sourceByListen := map[string]string{}
	for _, r := range rows {
		if r.Type != BlockTypeSource {
			continue
		}
		if listen := parseSourceListen(configs[r.ID]); listen != "" {
			sourceByListen[listen] = r.ID
		}
	}
	titles := map[string]string{}
	for _, r := range rows {
		if r.Type != BlockTypeLink {
			continue
		}
		host := parseLinkUpstreamHost(configs[r.ID])
		if src, ok := sourceByListen[host]; ok {
			titles[r.ID] = src + " → " + r.ID
		}
	}
	return titles
}

func parseSourceListen(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var sc struct {
		Listen string `json:"listen"`
	}
	if json.Unmarshal(raw, &sc) != nil {
		return ""
	}
	return sc.Listen
}

func parseLinkUpstreamHost(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var lc struct {
		Upstream string `json:"upstream"`
	}
	if json.Unmarshal(raw, &lc) != nil || lc.Upstream == "" {
		return ""
	}
	u, err := url.Parse(lc.Upstream)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Host
}
