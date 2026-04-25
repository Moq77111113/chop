// Package dashboard embeds the built Vite/Solid SPA into the chop binary.
package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

const distSubdir = "dist"

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded dashboard SPA.
// It expects `web/dist` contents to have been copied into this package's
// `dist/` directory at build time (see Makefile).
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, distSubdir)
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}
