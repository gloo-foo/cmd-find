package command_test

import (
	"context"
	"os"
	"sort"
	"testing"

	command "github.com/gloo-foo/cmd-find"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// tree builds a fixed in-memory filesystem used across the tests:
//
//	/root
//	/root/a.txt
//	/root/keep.go
//	/root/sub
//	/root/sub/b.txt
//	/root/sub/deep
//	/root/sub/deep/c.go
func tree(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	for _, f := range []string{"/root/a.txt", "/root/keep.go", "/root/sub/b.txt", "/root/sub/deep/c.go"} {
		if err := afero.WriteFile(fs, f, []byte(""), 0o644); err != nil {
			t.Fatalf("seed %s: %v", f, err)
		}
	}
	return fs
}

// run collects the sorted paths a Find source emits.
func run(t *testing.T, src gloo.Source[[]byte]) []string {
	t.Helper()
	got, err := gloo.Collect(context.Background(), src.Stream(context.Background()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := make([]string, len(got))
	for i, b := range got {
		out[i] = string(b)
	}
	sort.Strings(out)
	return out
}

func assertPaths(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestFind_WalksWholeTree(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t))))
	assertPaths(t, got, []string{
		"/root", "/root/a.txt", "/root/keep.go",
		"/root/sub", "/root/sub/b.txt", "/root/sub/deep", "/root/sub/deep/c.go",
	})
}

func TestFind_NameGlob(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindName("*.go")))
	assertPaths(t, got, []string{"/root/keep.go", "/root/sub/deep/c.go"})
}

func TestFind_NameGlobNoMatch(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindName("*.md")))
	assertPaths(t, got, nil)
}

func TestFind_MalformedGlobMatchesNothing(t *testing.T) {
	// filepath.Match returns ErrBadPattern for an unterminated class; find
	// treats that as "no match" rather than failing the walk.
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindName("[")))
	assertPaths(t, got, nil)
}

func TestFind_TypeFile(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindTypeFile))
	assertPaths(t, got, []string{
		"/root/a.txt", "/root/keep.go", "/root/sub/b.txt", "/root/sub/deep/c.go",
	})
}

func TestFind_TypeDir(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindTypeDir))
	assertPaths(t, got, []string{"/root", "/root/sub", "/root/sub/deep"})
}

func TestFind_TypeUnknownIsUnfiltered(t *testing.T) {
	// An unrecognized -type value imposes no restriction (every entry passes).
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindType("x")))
	assertPaths(t, got, []string{
		"/root", "/root/a.txt", "/root/keep.go",
		"/root/sub", "/root/sub/b.txt", "/root/sub/deep", "/root/sub/deep/c.go",
	})
}

func TestFind_MaxDepthZero(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindMaxDepth(0)))
	assertPaths(t, got, []string{"/root"})
}

func TestFind_MaxDepthOne(t *testing.T) {
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindMaxDepth(1)))
	assertPaths(t, got, []string{"/root", "/root/a.txt", "/root/keep.go", "/root/sub"})
}

func TestFind_MaxDepthPrunesDeepFile(t *testing.T) {
	// A file (non-directory) past the limit is skipped without ending the walk,
	// exercising the non-directory branch of prune.
	got := run(t, command.Find("/root", command.FindFs(tree(t)), command.FindMaxDepth(2)))
	assertPaths(t, got, []string{
		"/root", "/root/a.txt", "/root/keep.go",
		"/root/sub", "/root/sub/b.txt", "/root/sub/deep",
	})
}

func TestFind_TypeAndMaxDepthCompose(t *testing.T) {
	got := run(t, command.Find("/root",
		command.FindFs(tree(t)), command.FindTypeFile, command.FindMaxDepth(1)))
	assertPaths(t, got, []string{"/root/a.txt", "/root/keep.go"})
}

func TestFind_NameAndTypeCompose(t *testing.T) {
	got := run(t, command.Find("/root",
		command.FindFs(tree(t)), command.FindName("*.go"), command.FindTypeFile))
	assertPaths(t, got, []string{"/root/keep.go", "/root/sub/deep/c.go"})
}

func TestFind_MissingRootIsAnError(t *testing.T) {
	_, err := gloo.Collect(context.Background(),
		command.Find("/nope", command.FindFs(afero.NewMemMapFs())).Stream(context.Background()))
	if err == nil {
		t.Fatal("walking a missing root should surface an error")
	}
}

// statFailFs wraps an afero.Fs and fails Stat for one path. Because it embeds
// the afero.Fs interface (not the concrete *MemMapFs), it does not satisfy
// afero.Lstater, so afero.Walk falls back to Stat for every child — letting this
// wrapper inject a per-entry stat error on a deep entry.
type statFailFs struct {
	afero.Fs
	failOn string
}

func (f statFailFs) Stat(name string) (os.FileInfo, error) {
	if name == f.failOn {
		return nil, errStatBoom
	}
	return f.Fs.Stat(name)
}

const errStatBoom statErr = "stat boom"

type statErr string

func (e statErr) Error() string { return string(e) }

func TestFind_DeepEntryStatErrorIsSkipped(t *testing.T) {
	// A stat error on a non-root entry is non-fatal: GNU find reports it and
	// keeps walking. The faulty child is omitted; the rest of the tree is still
	// emitted, and the walk completes without surfacing an error.
	base := tree(t)
	fs := statFailFs{Fs: base, failOn: "/root/sub"}
	got := run(t, command.Find("/root", command.FindFs(fs)))
	for _, p := range got {
		if p == "/root/sub" {
			t.Fatalf("entry with a stat error should be skipped, got %v", got)
		}
	}
	if len(got) == 0 {
		t.Fatalf("walk should still emit the reachable entries, got %v", got)
	}
}

func TestFind_RootStatErrorIsFatal(t *testing.T) {
	// A stat error on the root itself is fatal: there is nothing to walk, so the
	// error is surfaced downstream.
	fs := statFailFs{Fs: tree(t), failOn: "/root"}
	_, err := gloo.Collect(context.Background(),
		command.Find("/root", command.FindFs(fs)).Stream(context.Background()))
	if err == nil {
		t.Fatal("a stat error on the root should surface as an error")
	}
}

func TestFind_DefaultFilesystemIsOS(t *testing.T) {
	// With no FindFs option the source walks the OS filesystem. Pointing it at a
	// guaranteed-missing path proves the default wiring without depending on any
	// real directory contents: a real OS walk of a missing root returns an error.
	_, err := gloo.Collect(context.Background(),
		command.Find("/this/path/does/not/exist/cmd-find").Stream(context.Background()))
	if err == nil {
		t.Fatal("default OS filesystem walk of a missing root should error")
	}
}
