package command

import (
	"github.com/spf13/afero"
)

// FindFs injects the filesystem a find run walks. Production omits it and the
// real OS filesystem is used; tests pass FindFs{Fs: afero.NewMemMapFs()} for
// hermetic, 100%-coverable runs. This is the dependency-injection seam for the
// whole module.
type FindFs struct{ Fs afero.Fs }

// FindName filters entries whose base name matches the glob pattern, GNU -name
// style (filepath.Match). The empty pattern matches every entry.
type FindName string

// FindType restricts results to a single entry kind (GNU -type): "f" for
// files, "d" for directories. Any other value imposes no restriction. Prefer
// the FindTypeFile / FindTypeDir constants.
type FindType string

const (
	// FindTypeFile keeps only regular files (-type f).
	FindTypeFile FindType = "f"
	// FindTypeDir keeps only directories (-type d).
	FindTypeDir FindType = "d"
)

// FindMaxDepth limits the walk to entries at most n levels below the root
// (GNU -maxdepth). The root itself is depth 0.
type FindMaxDepth int

// unlimitedDepth is the sentinel meaning "no depth limit".
const unlimitedDepth FindMaxDepth = -1

// flags holds the folded configuration of a find run. It is an immutable value
// assembled by fold before the walk begins.
type flags struct {
	fs         afero.Fs
	name       FindName
	typeFilter FindType
	maxDepth   FindMaxDepth
}

// with folds one option value into the flag set. Values of any other type are
// ignored: find's only positional argument is the root, taken separately.
func (f flags) with(o any) flags {
	switch v := o.(type) {
	case FindFs:
		f.fs = v.Fs
	case FindName:
		f.name = v
	case FindType:
		f.typeFilter = v
	case FindMaxDepth:
		f.maxDepth = v
	}
	return f
}

// fold collapses the Find option values onto the defaults: the OS filesystem
// and an unlimited walk depth (so FindMaxDepth(0) stays distinguishable from
// "no -maxdepth given").
func fold(opts []any) flags {
	f := flags{fs: afero.NewOsFs(), maxDepth: unlimitedDepth}
	for _, o := range opts {
		f = f.with(o)
	}
	return f
}
