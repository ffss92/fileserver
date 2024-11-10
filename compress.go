package fileserver

import (
	"compress/gzip"
	"net/http"
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
