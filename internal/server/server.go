package server

import (
	"net/http"
	"sync"

	"github.com/coffee-cup/peel/internal/embed"
	"github.com/coffee-cup/peel/internal/image"
)

type Server struct {
	mu      sync.RWMutex
	image   *image.Image
	ref     string
	loadErr error
	mux     *http.ServeMux
}

func New(ref string) *Server {
	s := &Server{ref: ref, mux: http.NewServeMux()}

	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/image", s.handleImage)
	s.mux.HandleFunc("GET /api/layers", s.handleLayers)
	s.mux.HandleFunc("GET /api/layers/{id}/tree", s.handleLayerTree)
	s.mux.HandleFunc("GET /api/layers/{id}/diff", s.handleLayerDiff)
	s.mux.HandleFunc("GET /api/files/{layer}/{path...}", s.handleFileContent)

	s.mux.Handle("/", embed.FileServer())

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) SetImage(img *image.Image) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.image = img
}

func (s *Server) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadErr = err
}

// requireImage returns the loaded image or writes an error response.
// Returns nil if the image is not yet available.
func (s *Server) requireImage(w http.ResponseWriter) *image.Image {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.loadErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"status": "error",
			"ref":    s.ref,
			"error":  s.loadErr.Error(),
		})
		return nil
	}
	if s.image == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "loading",
			"ref":    s.ref,
		})
		return nil
	}
	return s.image
}
