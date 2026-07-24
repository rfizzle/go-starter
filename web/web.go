// Package web exposes the framework-neutral static assets embedded in the
// service binary. The committed build/index.html keeps clean Go builds
// independent of a frontend toolchain; downstream projects may replace the
// build directory with their frontend output.
package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
)

// Embed every committed frontend build artifact. At least build/index.html
// must remain present so a clean clone compiles without running Node.
//
//go:embed build/*
var static embed.FS

// fsFunc is shorthand for constructing a http.FileSystem
// implementation
type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

// AssetHandler returns a http.Handler that will serve files from
// the Assets embed.FS. When locating a file, it will strip the given
// prefix from the request and prepend the root to the filesystem
func AssetHandler(prefix, root string) http.Handler {
	newStatic, err := fs.Sub(static, "build")
	if err != nil {
		log.Fatal(err)
	}

	handler := fsFunc(func(name string) (fs.File, error) {
		assetPath := path.Join(root, name)

		// Forward root requests to index
		if name == "/" || name == "" {
			return newStatic.Open("index.html")
		}

		return newStatic.Open(assetPath)
	})

	return http.StripPrefix(prefix, http.FileServer(http.FS(handler)))
}
