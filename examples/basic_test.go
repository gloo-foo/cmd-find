package find_test

import (
	"github.com/gloo-foo/framework/patterns"
	"github.com/spf13/afero"

	find "github.com/gloo-foo/cmd-find"
)

func ExampleFind() {
	// find . -name "*.go" -maxdepth 2 over a deterministic in-memory tree.
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "main.go", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "doc/notes.txt", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "pkg/util.go", []byte(""), 0o644)

	if err := patterns.Run(
		find.Find(".", find.FindFs{Fs: fs}, find.FindName("*.go"), find.FindMaxDepth(2)),
	); err != nil {
		panic(err)
	}
	// Output:
	// main.go
	// pkg/util.go
}
