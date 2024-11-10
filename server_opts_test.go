package fileserver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestWithETagFunc(t *testing.T) {
	h := New(os.DirFS("testdata"), WithETagFunc(func(_ io.Reader) (string, error) {
		return "foo", nil
	}))

	srv := httptest.NewServer(http.StripPrefix("/", h))
	client := srv.Client()

	res, err := client.Get(srv.URL + "/file.txt")
	if err != nil {
		t.Fatalf("failed to get file from server: %s", err)
	}
	defer res.Body.Close()

	etag := res.Header.Get("ETag")
	if etag != "foo" {
		t.Errorf("expected etag to be foo but got %s", etag)
	}
}

func TestWithErrorHandler(t *testing.T) {
	h := New(os.DirFS("testdata"), WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("teapot"))
	}))

	srv := httptest.NewServer(http.StripPrefix("/", h))
	client := srv.Client()

	res, err := client.Get(srv.URL + "/invalid")
	if err != nil {
		t.Fatalf("failed to get file from server: %s", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("unexpected error reading response %s", err)
	}
	if !bytes.Equal(body, []byte("teapot")) {
		t.Errorf("expected body to be teapot but got %s", body)
	}
	if res.StatusCode != http.StatusTeapot {
		t.Errorf("expected status to be 418 but got %d", res.StatusCode)
	}
}

func TestWithCacheControlFunc(t *testing.T) {
	type target struct {
		url      string
		expected string
	}
	testCases := []struct {
		name             string
		targets          []target
		cacheControlFunc CacheControlFunc
	}{
		{
			name: "nil func",
			targets: []target{
				{
					url:      "/file.txt",
					expected: "",
				},
			},
			cacheControlFunc: nil,
		},
		{
			name: "custom func (static)",
			targets: []target{
				{
					url:      "/file.txt",
					expected: "private, no-cache",
				},
			},
			cacheControlFunc: func(r *http.Request) string {
				return "private, no-cache"
			},
		},
		{
			name: "custom func (dynamic)",
			targets: []target{
				{
					url:      "/file.txt",
					expected: "no-cache",
				},
				{
					url:      "/subdir/subfile.txt",
					expected: "private, no-cache",
				},
			},
			cacheControlFunc: func(r *http.Request) string {
				path := r.URL.Path
				if strings.HasPrefix(path, "subdir") {
					return "private, no-cache"
				}
				return "no-cache"
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			h := New(os.DirFS("testdata"), WithCacheControlFunc(tt.cacheControlFunc))

			for _, target := range tt.targets {
				srv := httptest.NewServer(http.StripPrefix("/", h))

				client := srv.Client()
				res, err := client.Get(srv.URL + target.url)
				if err != nil {
					t.Fatalf("unexpected error making request: %s", err)
				}
				defer res.Body.Close()

				cacheControl := res.Header.Get("Cache-Control")
				if cacheControl != target.expected {
					t.Errorf("expected cache control header value to be %s but got %s", target.expected, cacheControl)
				}
			}
		})
	}
}
