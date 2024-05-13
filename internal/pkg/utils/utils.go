package utils

import (
	"io/fs"
	"net/http"
	"path/filepath"
)

func FindFilesWithExtension(root, extension string) []string {
	var found []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == extension {
			found = append(found, d.Name())
		}
		return nil
	})
	return found
}

// func AddHeaders(headers map[string]string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		for k, v := range headers {
// 			w.Header().Set(k, v)
// 		}
// 	}
// }

func AddHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Cache-Control", "no-cache, no-store")
		if h != nil {
			h.ServeHTTP(w, r)
		}
	}
}

// func AddHeaders(h http.Handler, headers map[string]string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		for k, v := range headers {
// 			w.Header().Set(k, v)
// 		}
// 		h.ServeHTTP(w, r)
// 	}
// }
