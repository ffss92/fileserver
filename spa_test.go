package fileserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServeSPA(t *testing.T) {
	spa := os.DirFS("testdata/spa")
	h := http.StripPrefix("/", ServeSPA(spa, "index.html"))

	index, err := os.ReadFile("testdata/spa/index.html")
	if err != nil {
		t.Fatal(err)
	}

	app, err := os.ReadFile("testdata/spa/assets/app.js")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name    string
		path    string
		content []byte
		status  int
	}{
		{
			name:    "root path",
			path:    "/", // Technically invalid, but will fallback
			content: index,
			status:  http.StatusOK,
		},
		{
			name:    "index.html",
			path:    "/index.html",
			content: index,
			status:  http.StatusOK,
		},
		{
			name:    "dir",
			path:    "/assets",
			content: index, // Expect fallback content
			status:  http.StatusOK,
		},
		{
			name:    "unknown",
			path:    "/bogus",
			content: index, // Expect fallback content
			status:  http.StatusOK,
		},
		{
			name:    "invalid",
			path:    "/../../hacked",
			content: index, // Expect fallback content
			status:  http.StatusOK,
		},
		{
			name:    "app.js",
			path:    "/assets/app.js",
			content: app, // Expect asset content
			status:  http.StatusOK,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			h.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("expected status to be 200 but got %d", w.Code)
			}

			body := w.Body.Bytes()
			if !bytes.Equal(tt.content, body) {
				t.Fatal("mismatched content")
			}
		})
	}

}
