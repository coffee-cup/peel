package image

import (
	"encoding/hex"
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	maxTextBytes   = 1 << 20 // 1MB
	maxBinaryBytes = 16 << 10 // 16KB
)

// Analyze extracts all metadata, builds filesystem trees, and computes diffs.
// The returned Image is immutable and safe for concurrent reads.
func Analyze(img v1.Image, ref string) (*Image, error) {
	cf, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("digest: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("layers: %w", err)
	}

	// Correlate history entries with content layers using two-pointer.
	// History may have more entries than layers (empty/metadata-only layers).
	history := cf.History
	var layerInfos []LayerInfo
	var emptyFlags []bool
	contentIdx := 0

	for i, h := range history {
		li := LayerInfo{
			Index:   i,
			Command: h.CreatedBy,
			Empty:   h.EmptyLayer,
		}
		if h.EmptyLayer {
			emptyFlags = append(emptyFlags, true)
		} else if contentIdx < len(layers) {
			diffID, _ := layers[contentIdx].DiffID()
			size, _ := layers[contentIdx].Size()
			li.DiffID = diffID.String()
			li.Size = size
			emptyFlags = append(emptyFlags, false)
			contentIdx++
		} else {
			li.Empty = true
			emptyFlags = append(emptyFlags, true)
		}
		layerInfos = append(layerInfos, li)
	}

	// If no history at all, synthesize from layers
	if len(history) == 0 {
		for i, l := range layers {
			diffID, _ := l.DiffID()
			size, _ := l.Size()
			layerInfos = append(layerInfos, LayerInfo{
				Index:  i,
				DiffID: diffID.String(),
				Size:   size,
			})
			emptyFlags = append(emptyFlags, false)
		}
	}

	trees, err := buildCumulativeTrees(layers, emptyFlags)
	if err != nil {
		return nil, fmt.Errorf("trees: %w", err)
	}

	// Compute diffs between consecutive cumulative trees
	diffs := make([][]DiffEntry, len(emptyFlags))
	for i := range emptyFlags {
		if i == 0 {
			diffs[i] = computeDiff(nil, trees[i])
		} else {
			diffs[i] = computeDiff(trees[i-1], trees[i])
		}
		if diffs[i] == nil {
			diffs[i] = []DiffEntry{}
		}
	}

	info := ImageInfo{
		Ref:        ref,
		Digest:     digest.String(),
		Arch:       cf.Architecture,
		OS:         cf.OS,
		LayerCount: len(layerInfos),
		Config: ImageConfig{
			Env:        cf.Config.Env,
			Entrypoint: cf.Config.Entrypoint,
			Cmd:        cf.Config.Cmd,
			WorkingDir: cf.Config.WorkingDir,
			User:       cf.Config.User,
			Labels:     cf.Config.Labels,
		},
	}

	return &Image{
		Info:   info,
		Layers: layerInfos,
		Trees:  trees,
		Diffs:  diffs,
		img:    img,
	}, nil
}

// ReadFile reads file content from the cumulative filesystem at the given layer.
// Searches backward through content layers to find the file.
func (im *Image) ReadFile(layerIdx int, path string) (*FileContent, error) {
	if layerIdx < 0 || layerIdx >= len(im.Layers) {
		return nil, fmt.Errorf("layer index %d out of range [0, %d)", layerIdx, len(im.Layers))
	}

	layers, err := im.img.Layers()
	if err != nil {
		return nil, fmt.Errorf("layers: %w", err)
	}

	emptyFlags := make([]bool, len(im.Layers))
	for i, l := range im.Layers {
		emptyFlags[i] = l.Empty
	}

	data, size, err := readFileFromLayer(layers, emptyFlags, layerIdx, path)
	if err != nil {
		return nil, err
	}

	binary := isBinary(data)
	fc := &FileContent{
		Path:     path,
		Size:     size,
		IsBinary: binary,
	}

	if binary {
		if len(data) > maxBinaryBytes {
			data = data[:maxBinaryBytes]
			fc.Truncated = true
		}
		fc.Content = hex.EncodeToString(data)
	} else {
		if len(data) > maxTextBytes {
			data = data[:maxTextBytes]
			fc.Truncated = true
		}
		fc.Content = string(data)
	}

	return fc, nil
}

// isBinary checks the first 8KB for null bytes.
func isBinary(data []byte) bool {
	check := data
	if len(check) > 8192 {
		check = check[:8192]
	}
	for _, b := range check {
		if b == 0 {
			return true
		}
	}
	return false
}
