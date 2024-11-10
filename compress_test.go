package fileserver

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The first two bytes are always going to be 1F and 8B. The third byte is for compression, which is usually 08 for "DEFLATE". See: RFC 1951.
// Ref: https://www.ietf.org/rfc/rfc1951.txt
func isGzip(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

func TestGzipCompression(t *testing.T) {
	msg := []byte("Hello World")

	compressed := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(compressed)
	defer gzWriter.Close()

	_, err := io.Copy(gzWriter, bytes.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzw := newGzipResponseWriter(w)
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = gzw.Write([]byte("Hello World"))
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("unexpected error reading response body: %s", err)
	}
	if !isGzip(body) || !bytes.Equal(compressed.Bytes(), body) {
		t.Error("expected body to be compressed with gzip")
	}
}
