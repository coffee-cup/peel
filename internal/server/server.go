package server

import (
	"encoding/json"
	"net/http"

	"github.com/coffee-cup/peel/internal/embed"
)

func New(dev bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)

	// Serve embedded frontend assets for non-API routes
	mux.Handle("/", embed.FileServer())

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
