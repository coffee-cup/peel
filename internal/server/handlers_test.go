package server

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/coffee-cup/peel/internal/image"
)

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	img := buildTestImage(t)
	analyzed, err := image.Analyze(img, "test:latest")
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(New(analyzed))
}

func TestHealth(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestImage(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/image")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var info image.ImageInfo
	json.NewDecoder(resp.Body).Decode(&info)
	if info.Ref != "test:latest" {
		t.Fatalf("expected test:latest, got %s", info.Ref)
	}
}

func TestLayers(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/layers")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var layers []image.LayerInfo
	json.NewDecoder(resp.Body).Decode(&layers)
	if len(layers) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(layers))
	}
}

func TestLayerTree(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/layers/0/tree")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var tree image.FileNode
	json.NewDecoder(resp.Body).Decode(&tree)
	if len(tree.Children) == 0 {
		t.Fatal("expected root to have children")
	}
}

func TestLayerDiff(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/layers/0/diff")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var diffs []image.DiffEntry
	json.NewDecoder(resp.Body).Decode(&diffs)
	if len(diffs) == 0 {
		t.Fatal("expected diff entries")
	}
}

func TestLayerTree_NotFound(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/layers/999/tree")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestFileContent(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/files/0/etc/hello")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var fc image.FileContent
	json.NewDecoder(resp.Body).Decode(&fc)
	if fc.Content != "hello\n" {
		t.Fatalf("expected hello\\n, got %q", fc.Content)
	}
}

func TestFileContent_NotFound(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/files/0/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// --- test image builder ---

type tarEntry struct {
	name     string
	typeflag byte
	data     []byte
	linkname string
}

func buildTestImage(t *testing.T) v1.Image {
	t.Helper()
	layer0 := buildTarLayer(t, []tarEntry{
		{name: "etc/", typeflag: tar.TypeDir},
		{name: "etc/hello", typeflag: tar.TypeReg, data: []byte("hello\n")},
		{name: "usr/", typeflag: tar.TypeDir},
		{name: "usr/bin/", typeflag: tar.TypeDir},
		{name: "usr/bin/app", typeflag: tar.TypeReg, data: []byte("#!/bin/sh\n")},
	})
	layer1 := buildTarLayer(t, []tarEntry{
		{name: "etc/hello", typeflag: tar.TypeReg, data: []byte("hello2\n")},
		{name: "var/", typeflag: tar.TypeDir},
		{name: "var/new", typeflag: tar.TypeReg, data: []byte("new\n")},
	})

	img, err := mutate.Append(empty.Image,
		mutate.Addendum{
			Layer:   layer0,
			History: v1.History{CreatedBy: "ADD . /", Created: v1.Time{Time: time.Unix(1, 0)}},
		},
		mutate.Addendum{
			Layer:   layer1,
			History: v1.History{CreatedBy: "COPY . /", Created: v1.Time{Time: time.Unix(2, 0)}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	cf, err := img.ConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	cf.Architecture = "amd64"
	cf.OS = "linux"
	cf.Config.Cmd = []string{"/app"}
	img, err = mutate.ConfigFile(img, cf)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func buildTarLayer(t *testing.T, entries []tarEntry) v1.Layer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeflag,
			Size:     int64(len(e.data)),
			Mode:     0644,
			Linkname: e.linkname,
		}
		if e.typeflag == tar.TypeDir {
			hdr.Mode = 0755
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if len(e.data) > 0 {
			if _, err := tw.Write(e.data); err != nil {
				t.Fatal(err)
			}
		}
	}
	tw.Close()
	layer, err := tarball.LayerFromReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	return layer
}
