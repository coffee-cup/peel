package image

import (
	"archive/tar"
	"runtime"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

func TestParsePlatform_Empty(t *testing.T) {
	p, err := ParsePlatform("")
	if err != nil {
		t.Fatal(err)
	}
	if p.OS != "linux" {
		t.Fatalf("expected linux, got %s", p.OS)
	}
	if p.Architecture != runtime.GOARCH {
		t.Fatalf("expected %s, got %s", runtime.GOARCH, p.Architecture)
	}
}

func TestParsePlatform_Valid(t *testing.T) {
	p, err := ParsePlatform("linux/arm64")
	if err != nil {
		t.Fatal(err)
	}
	if p.OS != "linux" || p.Architecture != "arm64" {
		t.Fatalf("got %s/%s", p.OS, p.Architecture)
	}
}

func TestParsePlatform_Invalid(t *testing.T) {
	_, err := ParsePlatform("badformat")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsBinary_Text(t *testing.T) {
	if isBinary([]byte("hello world")) {
		t.Fatal("text should not be binary")
	}
}

func TestIsBinary_Binary(t *testing.T) {
	if !isBinary([]byte{0x00, 0x01, 0x02}) {
		t.Fatal("null byte should be binary")
	}
}

func TestIsBinary_Empty(t *testing.T) {
	if isBinary([]byte{}) {
		t.Fatal("empty should not be binary")
	}
}

func TestIsBinary_NullPast8KB(t *testing.T) {
	data := make([]byte, 9000)
	for i := range data {
		data[i] = 'a'
	}
	data[8500] = 0x00
	if isBinary(data) {
		t.Fatal("null past 8KB should not be detected")
	}
}

func TestAnalyze_ReversedDaemonHistory(t *testing.T) {
	base := buildTarLayer(t, []tarEntry{
		{name: "base.txt", typeflag: tar.TypeReg, data: []byte("base\n")},
	})
	top := buildTarLayer(t, []tarEntry{
		{name: "top.txt", typeflag: tar.TypeReg, data: []byte("top\n")},
	})

	img, err := mutate.Append(empty.Image,
		mutate.Addendum{
			Layer:   base,
			History: v1.History{CreatedBy: "base-cmd", Created: v1.Time{Time: time.Unix(1, 0)}},
		},
		mutate.Addendum{
			Layer:   top,
			History: v1.History{CreatedBy: "top-cmd", Created: v1.Time{Time: time.Unix(2, 0)}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	wrapped := &reversedHistoryImage{img}
	result, err := Analyze(wrapped, "test:reversed")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Layers) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(result.Layers))
	}
	if result.Layers[0].Command != "base-cmd" {
		t.Errorf("layer 0 command: got %q, want %q", result.Layers[0].Command, "base-cmd")
	}
	if result.Layers[1].Command != "top-cmd" {
		t.Errorf("layer 1 command: got %q, want %q", result.Layers[1].Command, "top-cmd")
	}

	// Verify diff contents match layer order
	hasFile := func(diffs []DiffEntry, name string) bool {
		for _, d := range diffs {
			if d.Path == "/"+name {
				return true
			}
		}
		return false
	}
	if !hasFile(result.Diffs[0], "base.txt") {
		t.Errorf("layer 0 diff should contain /base.txt, got %v", result.Diffs[0])
	}
	if !hasFile(result.Diffs[1], "top.txt") {
		t.Errorf("layer 1 diff should contain /top.txt, got %v", result.Diffs[1])
	}
}

func TestAnalyze_ViaTestImage(t *testing.T) {
	img := testImage(t)

	// 3 history entries = 3 layers
	if len(img.Layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(img.Layers))
	}

	// Middle layer is empty
	if !img.Layers[1].Empty {
		t.Error("layer 1 should be empty")
	}
	if img.Layers[0].Empty || img.Layers[2].Empty {
		t.Error("layers 0 and 2 should not be empty")
	}

	// Content layers should have diffIDs
	if img.Layers[0].DiffID == "" {
		t.Error("layer 0 should have diffID")
	}
	if img.Layers[2].DiffID == "" {
		t.Error("layer 2 should have diffID")
	}

	// History correlation
	if img.Layers[0].Command != "ADD . /" {
		t.Errorf("unexpected command: %s", img.Layers[0].Command)
	}
	if img.Layers[1].Command != "ENV A=1" {
		t.Errorf("unexpected command: %s", img.Layers[1].Command)
	}

	// Image info
	if img.Info.Ref != "test:latest" {
		t.Errorf("unexpected ref: %s", img.Info.Ref)
	}
	if img.Info.Arch != "amd64" || img.Info.OS != "linux" {
		t.Errorf("unexpected platform: %s/%s", img.Info.OS, img.Info.Arch)
	}
}
