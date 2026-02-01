package embed

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var assets embed.FS

// FileServer returns an http.Handler that serves the embedded frontend assets.
func FileServer() http.Handler {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}
	return http.FileServerFS(sub)
}
