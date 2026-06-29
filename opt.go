package command

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// flags holds the parsed configuration of a find run. It is an immutable value
// assembled by the framework's Switch mechanism before the walk begins.
type flags struct {
	fs         afero.Fs
	name       namePattern
	typeFilter typeFilter
	maxDepth   maxDepth
}

// findFs injects the filesystem a find run walks. Production uses the real OS
// filesystem; tests inject afero.NewMemMapFs() for hermetic, 100%-coverable
// runs. This is the dependency-injection seam for the whole module.
type findFs struct{ afero.Fs }

// FindFs sets the filesystem to walk. Tests pass FindFs{Fs: afero.NewMemMapFs()};
// production omits it and the OS filesystem is used.
func FindFs(fs afero.Fs) gloo.Switch[flags] { return findFs{Fs: fs} }

func (f findFs) Configure(c *flags) { c.fs = f.Fs }

// namePattern is a glob (filepath.Match) tested against each entry's base name.
// The empty pattern matches every entry.
type namePattern string

// FindName filters entries whose base name matches the glob pattern (GNU -name).
func FindName(pattern string) gloo.Switch[flags] { return namePattern(pattern) }

func (n namePattern) Configure(c *flags) { c.name = n }

// typeFilter restricts results to a single entry kind (GNU -type).
type typeFilter string

const (
	// FindTypeFile keeps only regular files (-type f).
	FindTypeFile typeFilter = "f"
	// FindTypeDir keeps only directories (-type d).
	FindTypeDir typeFilter = "d"
)

// FindType filters entries by kind: "f" for files, "d" for directories. Any
// other value imposes no restriction. Prefer the FindTypeFile / FindTypeDir
// constants.
func FindType(kind string) gloo.Switch[flags] { return typeFilter(kind) }

func (t typeFilter) Configure(c *flags) { c.typeFilter = t }

// maxDepth limits how deep below the root the walk descends (GNU -maxdepth). A
// negative value (the zero default is remapped in Find) means unlimited.
type maxDepth int

// unlimitedDepth is the sentinel meaning "no depth limit".
const unlimitedDepth maxDepth = -1

// FindMaxDepth limits the walk to entries at most n levels below the root. The
// root itself is depth 0.
func FindMaxDepth(n int) gloo.Switch[flags] { return maxDepth(n) }

func (d maxDepth) Configure(c *flags) { c.maxDepth = d }
