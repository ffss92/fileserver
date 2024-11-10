//go:build go1.20
// +build go1.20

package fileserver

import "net/http"

func (gz *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return gz.ResponseWriter
}
