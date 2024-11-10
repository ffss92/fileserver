package fileserver

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

type ServerOptFn func(s *Server)

func WithETagFunc(etagFn ETagFunc) ServerOptFn {
	return func(s *Server) {
		s.etagFn = etagFn
	}
}

type Server struct {
	fs     fs.FS
	etagFn ETagFunc
}

func New(fs fs.FS, opts ...ServerOptFn) *Server {
	server := &Server{
		fs:     fs,
		etagFn: CalculateETag,
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

func Serve(dir string) http.Handler {
	return &Server{
		fs:     os.DirFS(dir),
		etagFn: CalculateETag,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fileName := r.URL.Path

	file, err := s.fs.Open(fileName)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrInvalid):
			http.Error(w, "invalid path", http.StatusBadRequest)
		case errors.Is(err, fs.ErrNotExist):
			http.NotFound(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	content := file.(io.ReadSeeker)

	// Calculate ETag
	if s.etagFn != nil {
		etag, err := s.etagFn(content)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if _, err := content.Seek(0, io.SeekStart); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("ETag", etag)
	}

	w.Header().Set("Vary", "Accept-Encoding")
	w.Header().Set("Cache-Control", "no-cache")

	// Compressed (gzip)
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gzw := newGzipResponseWriter(w)
		defer gzw.Close()

		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(gzw, r, fileName, stat.ModTime(), content)
		return
	}

	http.ServeContent(w, r, fileName, stat.ModTime(), content)
}
