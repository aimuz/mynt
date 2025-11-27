package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func spaHandler(filesystem fs.FS, index string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		_, err := fs.Stat(filesystem, name)
		if err != nil {
			http.ServeFileFS(w, r, filesystem, index)
			return
		}
		http.ServeFileFS(w, r, filesystem, name)
	}
}
