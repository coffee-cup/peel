package image

import (
	"archive/tar"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// buildLayerTree reads a layer's tar and builds a FileNode tree.
func buildLayerTree(layer v1.Layer) (*FileNode, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("uncompress layer: %w", err)
	}
	defer rc.Close()

	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir}
	lookup := map[string]*FileNode{"/": root}

	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		cleanPath := "/" + strings.TrimPrefix(path.Clean(hdr.Name), "/")

		var ft FileType
		switch hdr.Typeflag {
		case tar.TypeDir:
			ft = FileTypeDir
		case tar.TypeSymlink:
			ft = FileTypeSymlink
		case tar.TypeReg, tar.TypeRegA:
			ft = FileTypeFile
		default:
			continue
		}

		node := &FileNode{
			Name:       path.Base(cleanPath),
			Path:       cleanPath,
			Type:       ft,
			Size:       hdr.Size,
			LinkTarget: hdr.Linkname,
		}

		ensureParents(lookup, root, cleanPath)

		parent := lookup[path.Dir(cleanPath)]
		if existing, ok := lookup[cleanPath]; ok {
			// Replace existing node in parent's children
			for i, c := range parent.Children {
				if c == existing {
					parent.Children[i] = node
					break
				}
			}
		} else {
			parent.Children = append(parent.Children, node)
		}
		lookup[cleanPath] = node
	}

	sortTree(root)
	return root, nil
}

// ensureParents creates any missing ancestor directories for p.
func ensureParents(lookup map[string]*FileNode, root *FileNode, p string) {
	dir := path.Dir(p)
	if _, ok := lookup[dir]; ok {
		return
	}
	// Recursively ensure grandparents
	ensureParents(lookup, root, dir)
	node := &FileNode{Name: path.Base(dir), Path: dir, Type: FileTypeDir}
	parentDir := lookup[path.Dir(dir)]
	parentDir.Children = append(parentDir.Children, node)
	lookup[dir] = node
}

func sortTree(n *FileNode) {
	if len(n.Children) == 0 {
		return
	}
	sort.Slice(n.Children, func(i, j int) bool {
		a, b := n.Children[i], n.Children[j]
		if (a.Type == FileTypeDir) != (b.Type == FileTypeDir) {
			return a.Type == FileTypeDir
		}
		return a.Name < b.Name
	})
	for _, c := range n.Children {
		sortTree(c)
	}
}

// mergeTrees deep-copies base and applies overlay on top, handling whiteouts.
func mergeTrees(base, overlay *FileNode) *FileNode {
	merged := deepCopy(base)
	applyOverlay(merged, overlay)
	sortTree(merged)
	return merged
}

func deepCopy(n *FileNode) *FileNode {
	cp := *n
	if len(n.Children) > 0 {
		cp.Children = make([]*FileNode, len(n.Children))
		for i, c := range n.Children {
			cp.Children[i] = deepCopy(c)
		}
	}
	return &cp
}

func applyOverlay(base, overlay *FileNode) {
	if overlay == nil {
		return
	}

	baseLookup := make(map[string]*FileNode, len(base.Children))
	for _, c := range base.Children {
		baseLookup[c.Name] = c
	}

	// Check for opaque whiteout — clears all base children
	hasOpaque := false
	for _, c := range overlay.Children {
		if c.Name == ".wh..wh..opq" {
			hasOpaque = true
			break
		}
	}
	if hasOpaque {
		base.Children = nil
		baseLookup = nil
	}

	for _, oc := range overlay.Children {
		// Skip whiteout markers themselves
		if oc.Name == ".wh..wh..opq" {
			continue
		}

		// Individual whiteout: .wh.NAME means delete NAME
		if strings.HasPrefix(oc.Name, ".wh.") {
			target := strings.TrimPrefix(oc.Name, ".wh.")
			base.Children = removeChild(base.Children, target)
			delete(baseLookup, target)
			continue
		}

		existing, exists := baseLookup[oc.Name]
		if !exists {
			// New entry
			base.Children = append(base.Children, deepCopy(oc))
			continue
		}

		// Both are dirs — recurse
		if existing.Type == FileTypeDir && oc.Type == FileTypeDir {
			applyOverlay(existing, oc)
			continue
		}

		// Type mismatch or file→file replacement — overlay wins
		replacement := deepCopy(oc)
		for i, c := range base.Children {
			if c.Name == oc.Name {
				base.Children[i] = replacement
				break
			}
		}
	}
}

func removeChild(children []*FileNode, name string) []*FileNode {
	for i, c := range children {
		if c.Name == name {
			return append(children[:i], children[i+1:]...)
		}
	}
	return children
}

// buildCumulativeTrees builds the merged filesystem tree at each layer.
// Empty layers share the previous tree (safe because trees are immutable after construction).
func buildCumulativeTrees(layers []v1.Layer, emptyFlags []bool) ([]*FileNode, error) {
	trees := make([]*FileNode, len(emptyFlags))
	var prev *FileNode

	layerIdx := 0
	for i, empty := range emptyFlags {
		if empty {
			trees[i] = prev
			continue
		}
		if layerIdx >= len(layers) {
			trees[i] = prev
			continue
		}
		tree, err := buildLayerTree(layers[layerIdx])
		if err != nil {
			return nil, fmt.Errorf("layer %d: %w", i, err)
		}
		if prev == nil {
			prev = tree
		} else {
			prev = mergeTrees(prev, tree)
		}
		trees[i] = prev
		layerIdx++
	}
	return trees, nil
}

// computeDiff walks two trees and reports added/modified/deleted entries.
func computeDiff(prev, curr *FileNode) []DiffEntry {
	var diffs []DiffEntry
	if prev == nil && curr == nil {
		return diffs
	}
	if prev == nil {
		collectAll(curr, ChangeAdded, &diffs)
		return diffs
	}
	if curr == nil {
		collectAll(prev, ChangeDeleted, &diffs)
		return diffs
	}
	diffWalk(prev, curr, &diffs)
	return diffs
}

func diffWalk(prev, curr *FileNode, diffs *[]DiffEntry) {
	prevLookup := make(map[string]*FileNode, len(prev.Children))
	for _, c := range prev.Children {
		prevLookup[c.Name] = c
	}

	currLookup := make(map[string]*FileNode, len(curr.Children))
	for _, c := range curr.Children {
		currLookup[c.Name] = c
	}

	// Check for additions and modifications
	for _, cn := range curr.Children {
		pn, existed := prevLookup[cn.Name]
		if !existed {
			collectAll(cn, ChangeAdded, diffs)
			continue
		}
		if cn.Type != pn.Type || cn.Size != pn.Size || cn.LinkTarget != pn.LinkTarget {
			*diffs = append(*diffs, DiffEntry{
				Path:       cn.Path,
				Type:       cn.Type,
				ChangeKind: ChangeModified,
				Size:       cn.Size,
			})
		}
		if cn.Type == FileTypeDir && pn.Type == FileTypeDir {
			diffWalk(pn, cn, diffs)
		}
	}

	// Check for deletions
	for _, pn := range prev.Children {
		if _, exists := currLookup[pn.Name]; !exists {
			collectAll(pn, ChangeDeleted, diffs)
		}
	}
}

func collectAll(n *FileNode, kind ChangeKind, diffs *[]DiffEntry) {
	// Skip root itself
	if n.Path != "/" {
		*diffs = append(*diffs, DiffEntry{
			Path:       n.Path,
			Type:       n.Type,
			ChangeKind: kind,
			Size:       n.Size,
		})
	}
	for _, c := range n.Children {
		collectAll(c, kind, diffs)
	}
}

// buildPathLookup creates a map from path to FileNode for quick lookups.
func buildPathLookup(root *FileNode) map[string]*FileNode {
	lookup := make(map[string]*FileNode)
	var walk func(n *FileNode)
	walk = func(n *FileNode) {
		lookup[n.Path] = n
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return lookup
}

// resolveSymlink follows symlink chains in the tree, returning the resolved path.
// Returns the original path if not a symlink. Returns an error for dangling or cyclic links.
func resolveSymlink(root *FileNode, filePath string, maxHops int) (string, error) {
	if maxHops <= 0 {
		maxHops = 10
	}
	lookup := buildPathLookup(root)

	current := filePath
	for i := 0; i < maxHops; i++ {
		node, ok := lookup[current]
		if !ok {
			return "", fmt.Errorf("dangling symlink: %s not found", current)
		}
		if node.Type != FileTypeSymlink {
			return current, nil
		}
		target := node.LinkTarget
		if !strings.HasPrefix(target, "/") {
			target = path.Join(path.Dir(current), target)
		}
		target = "/" + strings.TrimPrefix(path.Clean(target), "/")
		current = target
	}
	return "", fmt.Errorf("symlink cycle: exceeded %d hops from %s", maxHops, filePath)
}

// readFileFromLayer searches backward through layers to find a file at path.
func readFileFromLayer(layers []v1.Layer, emptyFlags []bool, layerIdx int, filePath string) ([]byte, int64, error) {
	cleanPath := "/" + strings.TrimPrefix(path.Clean(filePath), "/")

	// Map layer index (including empty) to content layer indices
	contentIdx := -1
	for i := 0; i <= layerIdx; i++ {
		if !emptyFlags[i] {
			contentIdx++
		}
	}

	// Search backward through content layers
	for ci := contentIdx; ci >= 0; ci-- {
		if ci >= len(layers) {
			continue
		}
		data, size, err := searchLayerTar(layers[ci], cleanPath)
		if err != nil {
			return nil, 0, err
		}
		if data != nil {
			return data, size, nil
		}
	}
	return nil, 0, fmt.Errorf("file not found: %s", filePath)
}

func searchLayerTar(layer v1.Layer, filePath string) ([]byte, int64, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, 0, fmt.Errorf("uncompress: %w", err)
	}
	defer rc.Close()

	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, 0, nil
		}
		if err != nil {
			return nil, 0, fmt.Errorf("read tar: %w", err)
		}

		cleanName := "/" + strings.TrimPrefix(path.Clean(hdr.Name), "/")
		if cleanName != filePath {
			continue
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			return nil, 0, nil
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, 0, fmt.Errorf("read file: %w", err)
		}
		return data, hdr.Size, nil
	}
}
