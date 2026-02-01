package image

import v1 "github.com/google/go-containerregistry/pkg/v1"

type FileType string

const (
	FileTypeFile    FileType = "file"
	FileTypeDir     FileType = "dir"
	FileTypeSymlink FileType = "symlink"
)

type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeDeleted  ChangeKind = "deleted"
)

type ImageInfo struct {
	Ref        string      `json:"ref"`
	Digest     string      `json:"digest"`
	Arch       string      `json:"arch"`
	OS         string      `json:"os"`
	Config     ImageConfig `json:"config"`
	LayerCount int         `json:"layerCount"`
}

type ImageConfig struct {
	Env        []string          `json:"env"`
	Entrypoint []string          `json:"entrypoint"`
	Cmd        []string          `json:"cmd"`
	WorkingDir string            `json:"workingDir"`
	User       string            `json:"user"`
	Labels     map[string]string `json:"labels"`
}

type LayerInfo struct {
	Index   int    `json:"index"`
	DiffID  string `json:"diffID"`
	Size    int64  `json:"size"`
	Command string `json:"command"`
	Empty   bool   `json:"empty"`
}

type FileNode struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Type       FileType    `json:"type"`
	Size       int64       `json:"size"`
	LinkTarget string      `json:"linkTarget,omitempty"`
	Children   []*FileNode `json:"children,omitempty"`
}

type DiffEntry struct {
	Path       string     `json:"path"`
	Type       FileType   `json:"type"`
	ChangeKind ChangeKind `json:"changeKind"`
	Size       int64      `json:"size"`
}

type FileContent struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	IsBinary  bool   `json:"isBinary"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

// Image holds the fully-analyzed image in memory. Immutable after Analyze().
type Image struct {
	Info   ImageInfo     `json:"info"`
	Layers []LayerInfo   `json:"layers"`
	Trees  []*FileNode   // indexed by layer index; empty layers share previous tree
	Diffs  [][]DiffEntry // indexed by layer index
	img    v1.Image
}
