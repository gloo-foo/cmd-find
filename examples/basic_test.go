package find_test

import (
	"github.com/spf13/afero"

	find "github.com/gloo-foo/cmd-find"
	"github.com/gloo-foo/framework/patterns"
)

func ExampleFind() {
	// find . -name "*.go" -maxdepth 2 over a deterministic in-memory tree.
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "main.go", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "doc/notes.txt", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "pkg/util.go", []byte(""), 0o644)

	patterns.MustRun(
		find.Find(".", find.FindFs(fs), find.FindName("*.go"), find.FindMaxDepth(2)),
	)
	// Output:
	// main.go
	// pkg/util.go
}
