package server

import (
	"net/http"

	"github.com/coffee-cup/peel/internal/embed"
	"github.com/coffee-cup/peel/internal/image"
)

func New(img *image.Image) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/image", handleImage(img))
	mux.HandleFunc("GET /api/layers", handleLayers(img))
	mux.HandleFunc("GET /api/layers/{id}/tree", handleLayerTree(img))
	mux.HandleFunc("GET /api/layers/{id}/diff", handleLayerDiff(img))
	mux.HandleFunc("GET /api/files/{layer}/{path...}", handleFileContent(img))

	mux.Handle("/", embed.FileServer())

	return mux
}
