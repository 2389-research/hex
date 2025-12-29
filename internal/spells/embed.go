// ABOUTME: Embedded builtin spells filesystem
// ABOUTME: Uses go:embed to include builtin spells in the binary

package spells

import (
	"embed"
	"io/fs"
)

//go:embed builtin/*
var builtinFS embed.FS

// BuiltinFS returns the embedded filesystem containing builtin spells
func BuiltinFS() fs.FS {
	// Return the builtin subdirectory as the root
	sub, err := fs.Sub(builtinFS, "builtin")
	if err != nil {
		// This should never happen since builtin/ is embedded
		return builtinFS
	}
	return sub
}
