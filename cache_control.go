package fileserver

import (
	"net/http"
	"slices"
)

type CacheControlFunc func(r *http.Request) string

// Sets the value of 'no-cache' to the 'Cache-Control' header for all files.
func NoCache(_ *http.Request) string {
	return "no-cache"
}

// Sets the value of 'public, max-age=31536000, immutable' to the 'Cache-Control' header for all files. You can provide
// a ignore list that won't be treated as permanent.
//
// This option is suitable for assets bundled with Vite, Webpack, etc.
func Immutable(ignore ...string) CacheControlFunc {
	return func(r *http.Request) string {
		if slices.Contains(ignore, r.URL.Path) {
			return "no-cache"
		}
		return "public, max-age=31536000, immutable"
	}
}
