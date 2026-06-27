package alias_test

import (
	"context"
	"sort"
	"testing"

	find "github.com/gloo-foo/cmd-find/alias"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// The alias package re-exports the constructor and flag values under unprefixed
// names. A mis-wired re-export (Type bound to the wrong constant, Name bound to
// MaxDepth, Find bound to the wrong function) compiles cleanly, so only behavior
// can prove the wiring. Each test exercises one re-export and asserts the exact
// set of paths it must produce over a fixed in-memory tree.
//
//	/root
//	/root/a.txt
//	/root/keep.go
//	/root/sub
//	/root/sub/c.go
func tree(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	for _, f := range []string{"/root/a.txt", "/root/keep.go", "/root/sub/c.go"} {
		if err := afero.WriteFile(fs, f, []byte(""), 0o644); err != nil {
			t.Fatalf("seed %s: %v", f, err)
		}
	}
	return fs
}

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

func TestAlias_FindAndFsWalkWholeTree(t *testing.T) {
	got := run(t, find.Find("/root", find.Fs(tree(t))))
	assertPaths(t, got, []string{"/root", "/root/a.txt", "/root/keep.go", "/root/sub", "/root/sub/c.go"})
}

func TestAlias_NameFiltersByGlob(t *testing.T) {
	got := run(t, find.Find("/root", find.Fs(tree(t)), find.Name("*.go")))
	assertPaths(t, got, []string{"/root/keep.go", "/root/sub/c.go"})
}

func TestAlias_TypeFileKeepsOnlyFiles(t *testing.T) {
	got := run(t, find.Find("/root", find.Fs(tree(t)), find.TypeFile))
	assertPaths(t, got, []string{"/root/a.txt", "/root/keep.go", "/root/sub/c.go"})
}

func TestAlias_TypeDirKeepsOnlyDirs(t *testing.T) {
	got := run(t, find.Find("/root", find.Fs(tree(t)), find.TypeDir))
	assertPaths(t, got, []string{"/root", "/root/sub"})
}

func TestAlias_MaxDepthLimitsDescent(t *testing.T) {
	got := run(t, find.Find("/root", find.Fs(tree(t)), find.MaxDepth(1)))
	assertPaths(t, got, []string{"/root", "/root/a.txt", "/root/keep.go", "/root/sub"})
}
