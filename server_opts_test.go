package fileserver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
