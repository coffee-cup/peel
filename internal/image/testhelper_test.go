package image

import (
	"archive/tar"
	"bytes"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// testImage builds a synthetic two-layer OCI image and runs Analyze on it.
//
// Layer 0: /etc/hello (file "hello\n"), /usr/bin/ (dir), /usr/bin/app (file), /lib/link (symlink → /etc/hello)
// Layer 1: /etc/hello (modified "hello2\n"), /var/new (added), .wh.usr (whiteout deletes /usr)
// Config: history with 3 entries (layer0, empty, layer1), env=["A=1"], cmd=["/app"]
func testImage(t *testing.T) *Image {
	t.Helper()

	layer0 := buildTarLayer(t, []tarEntry{
		{name: "etc/", typeflag: tar.TypeDir},
		{name: "etc/hello", typeflag: tar.TypeReg, data: []byte("hello\n")},
		{name: "usr/", typeflag: tar.TypeDir},
		{name: "usr/bin/", typeflag: tar.TypeDir},
		{name: "usr/bin/app", typeflag: tar.TypeReg, data: []byte("#!/bin/sh\n")},
		{name: "lib/", typeflag: tar.TypeDir},
		{name: "lib/link", typeflag: tar.TypeSymlink, linkname: "/etc/hello"},
	})

	layer1 := buildTarLayer(t, []tarEntry{
		{name: "etc/hello", typeflag: tar.TypeReg, data: []byte("hello2\n")},
		{name: "var/", typeflag: tar.TypeDir},
		{name: "var/new", typeflag: tar.TypeReg, data: []byte("new\n")},
		{name: ".wh.usr", typeflag: tar.TypeReg},
	})

	img, err := mutate.Append(empty.Image,
		mutate.Addendum{
			Layer:   layer0,
			History: v1.History{CreatedBy: "ADD . /", Created: v1.Time{Time: time.Unix(1, 0)}},
		},
		mutate.Addendum{
			History: v1.History{CreatedBy: "ENV A=1", EmptyLayer: true, Created: v1.Time{Time: time.Unix(2, 0)}},
		},
		mutate.Addendum{
			Layer:   layer1,
			History: v1.History{CreatedBy: "COPY --from=0 / /", Created: v1.Time{Time: time.Unix(3, 0)}},
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
	cf.Config.Env = []string{"A=1"}
	cf.Config.Cmd = []string{"/app"}
	img, err = mutate.ConfigFile(img, cf)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Analyze(img, "test:latest")
	if err != nil {
		t.Fatal(err)
	}
	return result
}

type tarEntry struct {
	name     string
	typeflag byte
	data     []byte
	linkname string
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
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	layer, err := tarball.LayerFromReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	return layer
}
