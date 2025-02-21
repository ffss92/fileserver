package fileserver

import (
	"net/http"
	"strings"
)

// Checks if the request accepts gzip encoded responses.
func acceptsGzip(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "gzip")
}
