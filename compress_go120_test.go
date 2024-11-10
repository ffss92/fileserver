package fileserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipWriterUnwrap(t *testing.T) {
	w := httptest.NewRecorder()
	gzw := newGzipResponseWriter(w)

	rc := http.NewResponseController(gzw)
	err := rc.Flush()
	if err != nil {
		t.Fatalf("unexpected error calling flush: %s", err)
	}
}
