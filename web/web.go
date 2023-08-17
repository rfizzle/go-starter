package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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
		if strings.HasPrefix(name, "api") {
			return nil, fs.ErrNotExist
		}

		assetPath := path.Join(root, name)

		// If we can't find the asset, return the default index.html
		// content
		f, err := newStatic.Open(assetPath)
		if os.IsNotExist(err) {
			return newStatic.Open("index.html")
		}

		// Otherwise assume this is a legitimate request routed
		// correctly
		return f, err
	})

	return http.StripPrefix(prefix, http.FileServer(http.FS(handler)))
}
