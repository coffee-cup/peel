package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/coffee-cup/peel/internal/image"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := "loading"
	if s.loadErr != nil {
		status = "error"
	} else if s.image != nil {
		status = "ready"
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": status, "ref": s.ref})
}

func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	img := s.requireImage(w)
	if img == nil {
		return
	}
	writeJSON(w, http.StatusOK, img.Info)
}

func (s *Server) handleLayers(w http.ResponseWriter, r *http.Request) {
	img := s.requireImage(w)
	if img == nil {
		return
	}
	writeJSON(w, http.StatusOK, img.Layers)
}

func (s *Server) handleLayerTree(w http.ResponseWriter, r *http.Request) {
	img := s.requireImage(w)
	if img == nil {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid layer id")
		return
	}
	if id < 0 || id >= len(img.Trees) {
		writeError(w, http.StatusNotFound, "layer not found")
		return
	}
	tree := img.Trees[id]
	if tree == nil {
		writeJSON(w, http.StatusOK, &image.FileNode{
			Name: "/", Path: "/", Type: image.FileTypeDir,
		})
		return
	}
	writeJSON(w, http.StatusOK, tree)
}

func (s *Server) handleLayerDiff(w http.ResponseWriter, r *http.Request) {
	img := s.requireImage(w)
	if img == nil {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid layer id")
		return
	}
	if id < 0 || id >= len(img.Diffs) {
		writeError(w, http.StatusNotFound, "layer not found")
		return
	}
	writeJSON(w, http.StatusOK, img.Diffs[id])
}

func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	img := s.requireImage(w)
	if img == nil {
		return
	}
	layer, err := strconv.Atoi(r.PathValue("layer"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid layer id")
		return
	}
	filePath := "/" + r.PathValue("path")

	fc, err := img.ReadFile(layer, filePath)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fc)
}
