package command

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// Find returns a Source that walks the tree rooted at path on the configured
// filesystem and emits the path of every entry that passes the active filters.
//
// Options (each a Switch over the run's flags):
//   - FindFs(fs):        filesystem to walk (default: the OS filesystem)
//   - FindName(glob):    keep entries whose base name matches glob (GNU -name)
//   - FindType("f"/"d"): keep only files or only directories (GNU -type)
//   - FindMaxDepth(n):   descend at most n levels below the root (GNU -maxdepth)
func Find(path string, opts ...gloo.Switch[flags]) gloo.Source[[]byte] {
	cfg := defaults()
	for _, o := range opts {
		o.Configure(&cfg)
	}
	return findSource{root: path, cfg: cfg}
}

// defaults seeds a flags value before any Switch is applied: the OS filesystem
// and an unlimited walk depth (so FindMaxDepth(0) stays distinguishable from
// "no -maxdepth given").
func defaults() flags {
	return flags{fs: afero.NewOsFs(), maxDepth: unlimitedDepth}
}

// findSource is an immutable Source: a root path plus the resolved flags. Value
// receiver throughout — it carries no mutable state.
type findSource struct {
	root string
	cfg  flags
}

// Stream walks the tree and emits each matching path. Walk errors on individual
// entries are skipped (GNU find reports and continues); a fatal walk error is
// forwarded downstream.
func (s findSource) Stream(ctx context.Context) gloo.Stream[[]byte] {
	return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, sendErr func(error)) {
		if err := afero.Walk(s.cfg.fs, s.root, s.visit(send)); err != nil {
			sendErr(err)
		}
	})
}

// visit builds the afero.WalkFunc for one walk: it applies the active filters to
// each entry and emits the path of those that pass.
func (s findSource) visit(send func([]byte) bool) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err == nil {
			return s.handle(path, info, send)
		}
		// An error on the root is fatal: there is nothing to walk, so surface it
		// (afero.Walk returns whatever we return here). An error on a deeper
		// entry is non-fatal — GNU find reports it and keeps walking — so we skip
		// just that entry by returning nil.
		if path == s.root {
			return err
		}
		return nil
	}
}

// handle applies the depth and filter rules to one readable entry, emitting its
// path when it passes. It returns the WalkFunc control value: filepath.SkipDir
// to stop descending a directory that lies past the depth limit, otherwise nil
// (a non-directory past the limit is simply not emitted; the walk continues).
func (s findSource) handle(path string, info fs.FileInfo, send func([]byte) bool) error {
	within := s.withinDepth(path)
	if !within && info.IsDir() {
		return filepath.SkipDir
	}
	if within && s.cfg.keeps(info) {
		send([]byte(path))
	}
	return nil
}

// withinDepth reports whether path lies at or above the configured -maxdepth
// limit. An unlimited walk admits every entry.
func (s findSource) withinDepth(path string) bool {
	if s.cfg.maxDepth == unlimitedDepth {
		return true
	}
	return depth(s.root, path) <= int(s.cfg.maxDepth)
}

// keeps reports whether an entry survives the type and name filters.
func (c flags) keeps(info fs.FileInfo) bool {
	return c.typeFilter.accepts(info) && c.name.matches(info.Name())
}

// accepts reports whether info's kind satisfies the type filter.
func (t typeFilter) accepts(info fs.FileInfo) bool {
	switch t {
	case FindTypeFile:
		return !info.IsDir()
	case FindTypeDir:
		return info.IsDir()
	default: // no restriction (the empty pattern, or an unknown -type value)
		return true
	}
}

// matches reports whether base satisfies the name glob. The empty pattern
// matches everything. A malformed pattern matches nothing (filepath.Match's
// ErrBadPattern is treated as no match, like GNU find ignoring it).
func (n namePattern) matches(base string) bool {
	if n == "" {
		return true
	}
	ok, _ := filepath.Match(string(n), base)
	return ok
}

// depth counts how many levels path lies below root. The root is depth 0, a
// direct child is depth 1, and so on.
func depth(root, path string) int {
	rel := strings.TrimPrefix(path, root)
	rel = strings.TrimPrefix(rel, string(filepath.Separator))
	if rel == "" {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}
