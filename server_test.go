package fileserver

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var errorETagFunc = ETagFunc(func(r io.Reader) (string, error) {
	return "", errors.New("something went wrong...")
})

func TestServer(t *testing.T) {
	h := New(os.DirFS("testdata"))
	srv := httptest.NewServer(http.StripPrefix("/static/", h))

	testCases := []struct {
		name         string
		status       int
		uncompressed bool
		etagFunc     ETagFunc
		newRequest   func() (*http.Request, error)
	}{
		{
			name:   "valid file",
			status: http.StatusOK,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/static/file.txt", nil)
			},
		},
		{
			name:   "valid file (uncompressed)",
			status: http.StatusOK,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/static/file.txt", nil)
			},
			uncompressed: true,
		},
		{
			name:   "invalid method",
			status: http.StatusMethodNotAllowed,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodPost, srv.URL+"/static/file.txt", nil)
			},
		},
		{
			name:   "invalid file (dont exist)",
			status: http.StatusNotFound,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/static/foo.txt", nil)
			},
		},
		{
			name:   "invalid file (subdir)",
			status: http.StatusNotFound,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/static/subdir", nil)
			},
		},
		{
			name:   "valid file (subdir)",
			status: http.StatusOK,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/static/subdir/subfile.txt", nil)
			},
		},
		{
			name:     "bad etag func",
			status:   http.StatusInternalServerError,
			etagFunc: errorETagFunc,
			newRequest: func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/static/file.txt", nil)
			},
		},
		{
			name:   "invalid path",
			status: http.StatusBadRequest,
			newRequest: func() (*http.Request, error) {
				// The behavior using the http.ServeMux here will be a 404, since the path
				// will end up being /server.go
				return http.NewRequest(http.MethodGet, srv.URL+"/static/../server.go", nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			client := srv.Client()
			if tt.uncompressed {
				// Disable default compression behavior
				client.Transport = &http.Transport{
					DisableCompression: true,
				}
			}

			// Set etag func
			if tt.etagFunc == nil {
				tt.etagFunc = CalculateETag
			}
			h.etagFn = tt.etagFunc
			t.Cleanup(func() {
				h.etagFn = CalculateETag
			})

			req, err := tt.newRequest()
			if err != nil {
				t.Fatalf("unexpected error creating request: %s", err)
			}

			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("unexpected error making request: %s", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.status {
				t.Errorf("expected status to be %d but got %d", tt.status, res.StatusCode)
			}

			// Success assertions
			if tt.status == 200 {
				vary := res.Header.Get("Vary")
				etag := res.Header.Get("ETag")
				contentEncoding := res.Header.Get("Content-Encoding")

				if vary != "Accept-Encoding" {
					t.Errorf("expected Vary header to include Accept-Encoding value")
				}
				if h.etagFn != nil && etag == "" {
					t.Error("expected etag to be set")
				}
				if tt.uncompressed && contentEncoding == "gzip" {
					t.Error("expected Content-Encoding to be empty")
				}
			}
		})
	}
}

func TestServe(t *testing.T) {
	h := Serve("testdata")
	srv := httptest.NewServer(http.StripPrefix("/", h))

	res, err := srv.Client().Get(srv.URL + "/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status to be 200 but got %d", res.StatusCode)
	}
}
