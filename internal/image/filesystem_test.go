package image

import (
	"testing"
)

// --- mergeTrees ---

func TestMergeTrees_OpaqueWhiteout(t *testing.T) {
	base := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "a", Path: "/a", Type: FileTypeFile, Size: 1},
		{Name: "b", Path: "/b", Type: FileTypeFile, Size: 2},
	}}
	overlay := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: ".wh..wh..opq", Path: "/.wh..wh..opq", Type: FileTypeFile},
		{Name: "c", Path: "/c", Type: FileTypeFile, Size: 3},
	}}
	merged := mergeTrees(base, overlay)
	if len(merged.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(merged.Children))
	}
	if merged.Children[0].Name != "c" {
		t.Fatalf("expected child c, got %s", merged.Children[0].Name)
	}
}

func TestMergeTrees_IndividualWhiteout(t *testing.T) {
	base := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "keep", Path: "/keep", Type: FileTypeFile},
		{Name: "remove", Path: "/remove", Type: FileTypeFile},
	}}
	overlay := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: ".wh.remove", Path: "/.wh.remove", Type: FileTypeFile},
	}}
	merged := mergeTrees(base, overlay)
	if len(merged.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(merged.Children))
	}
	if merged.Children[0].Name != "keep" {
		t.Fatalf("expected keep, got %s", merged.Children[0].Name)
	}
}

func TestMergeTrees_DirDirMerge(t *testing.T) {
	base := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "dir", Path: "/dir", Type: FileTypeDir, Children: []*FileNode{
			{Name: "old", Path: "/dir/old", Type: FileTypeFile},
		}},
	}}
	overlay := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "dir", Path: "/dir", Type: FileTypeDir, Children: []*FileNode{
			{Name: "new", Path: "/dir/new", Type: FileTypeFile},
		}},
	}}
	merged := mergeTrees(base, overlay)
	dir := merged.Children[0]
	if len(dir.Children) != 2 {
		t.Fatalf("expected 2 children in dir, got %d", len(dir.Children))
	}
}

func TestMergeTrees_TypeReplacement(t *testing.T) {
	base := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "x", Path: "/x", Type: FileTypeDir, Children: []*FileNode{
			{Name: "child", Path: "/x/child", Type: FileTypeFile},
		}},
	}}
	overlay := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "x", Path: "/x", Type: FileTypeFile, Size: 99},
	}}
	merged := mergeTrees(base, overlay)
	if merged.Children[0].Type != FileTypeFile {
		t.Fatalf("expected file, got %s", merged.Children[0].Type)
	}
	if merged.Children[0].Size != 99 {
		t.Fatalf("expected size 99, got %d", merged.Children[0].Size)
	}
}

func TestMergeTrees_NewFile(t *testing.T) {
	base := &FileNode{Name: "/", Path: "/", Type: FileTypeDir}
	overlay := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "new", Path: "/new", Type: FileTypeFile, Size: 10},
	}}
	merged := mergeTrees(base, overlay)
	if len(merged.Children) != 1 || merged.Children[0].Name != "new" {
		t.Fatal("expected new file added")
	}
}

// --- computeDiff ---

func TestComputeDiff_NilPrev(t *testing.T) {
	curr := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "a", Path: "/a", Type: FileTypeFile, Size: 1},
	}}
	diffs := computeDiff(nil, curr)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].ChangeKind != ChangeAdded {
		t.Fatalf("expected added, got %s", diffs[0].ChangeKind)
	}
}

func TestComputeDiff_AddedModifiedDeleted(t *testing.T) {
	prev := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "same", Path: "/same", Type: FileTypeFile, Size: 10},
		{Name: "mod", Path: "/mod", Type: FileTypeFile, Size: 5},
		{Name: "gone", Path: "/gone", Type: FileTypeFile, Size: 3},
	}}
	curr := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "same", Path: "/same", Type: FileTypeFile, Size: 10},
		{Name: "mod", Path: "/mod", Type: FileTypeFile, Size: 99},
		{Name: "new", Path: "/new", Type: FileTypeFile, Size: 7},
	}}
	diffs := computeDiff(prev, curr)
	kinds := map[ChangeKind]int{}
	for _, d := range diffs {
		kinds[d.ChangeKind]++
	}
	if kinds[ChangeAdded] != 1 || kinds[ChangeModified] != 1 || kinds[ChangeDeleted] != 1 {
		t.Fatalf("unexpected diff counts: %v", kinds)
	}
}

func TestComputeDiff_NestedDirs(t *testing.T) {
	prev := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "d", Path: "/d", Type: FileTypeDir, Children: []*FileNode{
			{Name: "f", Path: "/d/f", Type: FileTypeFile, Size: 1},
		}},
	}}
	curr := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "d", Path: "/d", Type: FileTypeDir, Children: []*FileNode{
			{Name: "f", Path: "/d/f", Type: FileTypeFile, Size: 1},
			{Name: "g", Path: "/d/g", Type: FileTypeFile, Size: 2},
		}},
	}}
	diffs := computeDiff(prev, curr)
	if len(diffs) != 1 || diffs[0].Path != "/d/g" {
		t.Fatalf("expected 1 added /d/g, got %v", diffs)
	}
}

func TestComputeDiff_NoChanges(t *testing.T) {
	tree := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "x", Path: "/x", Type: FileTypeFile, Size: 1},
	}}
	diffs := computeDiff(tree, tree)
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs, got %d", len(diffs))
	}
}

// --- resolveSymlink ---

func TestResolveSymlink_Absolute(t *testing.T) {
	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "link", Path: "/link", Type: FileTypeSymlink, LinkTarget: "/target"},
		{Name: "target", Path: "/target", Type: FileTypeFile},
	}}
	resolved, err := resolveSymlink(root, "/link", 10)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "/target" {
		t.Fatalf("expected /target, got %s", resolved)
	}
}

func TestResolveSymlink_Relative(t *testing.T) {
	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "dir", Path: "/dir", Type: FileTypeDir, Children: []*FileNode{
			{Name: "link", Path: "/dir/link", Type: FileTypeSymlink, LinkTarget: "../file"},
		}},
		{Name: "file", Path: "/file", Type: FileTypeFile},
	}}
	resolved, err := resolveSymlink(root, "/dir/link", 10)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "/file" {
		t.Fatalf("expected /file, got %s", resolved)
	}
}

func TestResolveSymlink_Chain(t *testing.T) {
	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "a", Path: "/a", Type: FileTypeSymlink, LinkTarget: "/b"},
		{Name: "b", Path: "/b", Type: FileTypeSymlink, LinkTarget: "/c"},
		{Name: "c", Path: "/c", Type: FileTypeFile},
	}}
	resolved, err := resolveSymlink(root, "/a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "/c" {
		t.Fatalf("expected /c, got %s", resolved)
	}
}

func TestResolveSymlink_Cycle(t *testing.T) {
	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "a", Path: "/a", Type: FileTypeSymlink, LinkTarget: "/b"},
		{Name: "b", Path: "/b", Type: FileTypeSymlink, LinkTarget: "/a"},
	}}
	_, err := resolveSymlink(root, "/a", 10)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestResolveSymlink_Dangling(t *testing.T) {
	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "link", Path: "/link", Type: FileTypeSymlink, LinkTarget: "/nope"},
	}}
	_, err := resolveSymlink(root, "/link", 10)
	if err == nil {
		t.Fatal("expected dangling error")
	}
}

func TestResolveSymlink_NonSymlink(t *testing.T) {
	root := &FileNode{Name: "/", Path: "/", Type: FileTypeDir, Children: []*FileNode{
		{Name: "file", Path: "/file", Type: FileTypeFile},
	}}
	resolved, err := resolveSymlink(root, "/file", 10)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "/file" {
		t.Fatalf("expected /file, got %s", resolved)
	}
}

// --- integration via testImage ---

func TestBuildCumulativeTrees(t *testing.T) {
	img := testImage(t)

	// 3 history entries: layer0, empty, layer1
	if len(img.Trees) != 3 {
		t.Fatalf("expected 3 trees, got %d", len(img.Trees))
	}

	// Tree at layer 0 should have /etc/hello, /usr/bin/app, /lib/link
	tree0 := img.Trees[0]
	lookup0 := buildPathLookup(tree0)
	for _, p := range []string{"/etc/hello", "/usr/bin/app", "/lib/link"} {
		if _, ok := lookup0[p]; !ok {
			t.Errorf("tree0 missing %s", p)
		}
	}

	// Tree at layer 1 (empty) shares tree0
	if img.Trees[1] != img.Trees[0] {
		t.Error("empty layer should share previous tree pointer")
	}

	// Tree at layer 2 should have /etc/hello (modified), /var/new; /usr should be gone
	tree2 := img.Trees[2]
	lookup2 := buildPathLookup(tree2)
	if _, ok := lookup2["/var/new"]; !ok {
		t.Error("tree2 missing /var/new")
	}
	if _, ok := lookup2["/usr"]; ok {
		t.Error("tree2 should not have /usr (whiteout)")
	}
	if _, ok := lookup2["/etc/hello"]; !ok {
		t.Error("tree2 missing /etc/hello")
	}
}

func TestReadFileFromLayer_Known(t *testing.T) {
	img := testImage(t)
	fc, err := img.ReadFile(0, "/etc/hello")
	if err != nil {
		t.Fatal(err)
	}
	if fc.Content != "hello\n" {
		t.Fatalf("expected hello\\n, got %q", fc.Content)
	}
}

func TestReadFileFromLayer_Modified(t *testing.T) {
	img := testImage(t)
	// Layer index 2 is the second content layer
	fc, err := img.ReadFile(2, "/etc/hello")
	if err != nil {
		t.Fatal(err)
	}
	if fc.Content != "hello2\n" {
		t.Fatalf("expected hello2\\n, got %q", fc.Content)
	}
}

func TestReadFileFromLayer_Missing(t *testing.T) {
	img := testImage(t)
	_, err := img.ReadFile(0, "/nonexistent")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
