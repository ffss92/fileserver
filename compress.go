package fileserver

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (gz *gzipResponseWriter) Write(b []byte) (int, error) {
	return gz.writer.Write(b)
}

func (gz *gzipResponseWriter) Close() error {
	return gz.writer.Close()
}

func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	gzw := gzip.NewWriter(w)
	return &gzipResponseWriter{
		ResponseWriter: w,
		writer:         gzw,
	}
}

// Checks if the request accepts gzip encoded responses.
func acceptsGzip(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "gzip")
}
