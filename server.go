package fileserver

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

var (
	ErrFileNotFound  = fmt.Errorf("file not found: %w", fs.ErrNotExist)
	ErrInvalidPath   = fmt.Errorf("invalid file path: %w", fs.ErrInvalid)
	ErrInvalidMethod = errors.New("invalid http method")
)

type Server struct {
	fs         fs.FS
	etagFn     ETagFunc
	errHandler func(w http.ResponseWriter, r *http.Request, err error)
}

func New(fs fs.FS, opts ...ServerOptFn) *Server {
	server := &Server{
		fs:         fs,
		etagFn:     CalculateETag,
		errHandler: defaultErrorHandler,
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

func Serve(dir string) http.Handler {
	return New(os.DirFS(dir))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errHandler(w, r, ErrInvalidMethod)
		return
	}

	fileName := r.URL.Path
	if fileName == "" {
		s.errHandler(w, r, ErrFileNotFound)
		return
	}

	file, err := s.fs.Open(fileName)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrInvalid):
			s.errHandler(w, r, ErrInvalidPath)
		case errors.Is(err, fs.ErrNotExist):
			s.errHandler(w, r, ErrFileNotFound)
		default:
			s.errHandler(w, r, fmt.Errorf("failed to open file: %w", err))
		}
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		s.errHandler(w, r, fmt.Errorf("failed to stat file: %w", err))
		return
	}

	if stat.IsDir() {
		s.errHandler(w, r, ErrFileNotFound)
		return
	}

	content := file.(io.ReadSeeker)

	// Calculate ETag
	if s.etagFn != nil {
		etag, err := s.etagFn(content)
		if err != nil {
			s.errHandler(w, r, fmt.Errorf("failed to calculate etag: %w", err))
			return
		}
		if _, err := content.Seek(0, io.SeekStart); err != nil {
			s.errHandler(w, r, fmt.Errorf("failed to seek content: %w", err))
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
