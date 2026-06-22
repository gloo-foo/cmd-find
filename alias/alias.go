// Package alias provides unprefixed names for the find command and its flags.
//
//	import find "github.com/gloo-foo/cmd-find/alias"
//	find.Find("/root", find.Name("*.go"), find.TypeFile)
package alias

import command "github.com/gloo-foo/cmd-find"

// Find re-exports the constructor.
var Find = command.Find

// Fs injects the filesystem to walk (the dependency-injection seam).
var Fs = command.FindFs

// Name filters entries whose base name matches the glob (-name).
var Name = command.FindName

// MaxDepth limits the walk depth relative to the root (-maxdepth).
var MaxDepth = command.FindMaxDepth

// TypeFile keeps only regular files (-type f).
const TypeFile = command.FindTypeFile

// TypeDir keeps only directories (-type d).
const TypeDir = command.FindTypeDir
