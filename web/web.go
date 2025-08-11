package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
)

// Add all the files in static, including hidden files.
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
