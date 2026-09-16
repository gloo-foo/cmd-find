// Package alias provides unprefixed names for the find command and its flags.
//
//	import find "github.com/gloo-foo/cmd-find/alias"
//	find.Find("/root", find.Name("*.go"), find.TypeFile)
package alias

import (
	gloo "github.com/gloo-foo/framework"

	command "github.com/gloo-foo/cmd-find"
)

// Find walks the tree rooted at root and emits every matching path; see the
// command package for the option set.
func Find(root gloo.File, opts ...any) gloo.Source[[]byte] { return command.Find(root, opts...) }

// Fs injects the filesystem to walk (the dependency-injection seam).
type Fs = command.FindFs

// Name filters entries whose base name matches the glob (-name).
type Name = command.FindName

// MaxDepth limits the walk depth relative to the root (-maxdepth).
type MaxDepth = command.FindMaxDepth

// TypeFile keeps only regular files (-type f).
const TypeFile = command.FindTypeFile

// TypeDir keeps only directories (-type d).
const TypeDir = command.FindTypeDir
