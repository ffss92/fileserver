package fileserver

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
)

var _ http.Handler = (*Server)(nil)

// Server implements an [http.Handler] for serving static files.
//
// It sets the ETag e Last-Modified headers and properly handles
// Range, If-Range, If-Match, If-None-Match, If-Modified-Since
// and If-Unmodified-Since through the use of [http.ServeContent].
//
// By default, ETag generation is done by md5 hashing the file contents
// and Cache-Control is set to 'no-cache'. This behavior is configurable
// by creating a new File Server using [fileserver.New] and providing the
// desired [fileserver.ServerOptFn] functional options.
type Server struct {
	fs             fs.FS
	etagFn         ETagFunc
	errHandler     ErrorHandlerFunc
	cacheControlFn CacheControlFunc
}

// Creates a new [Server]. It can be configured using functional options.
//
//	fileServer := New(myFS, WithErrorHandler(myErrorHandlerFunc))
//	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
func New(fs fs.FS, opts ...ServerOptFn) *Server {
	server := &Server{
		fs:             fs,
		etagFn:         calculateETag,
		errHandler:     defaultErrorHandler,
		cacheControlFn: noCacheAll,
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

// Creates a new file server for a given [fs.FS].
func ServeFS(fs fs.FS) http.Handler {
	return New(fs)
}

// Creates a new file server a dir using [os.DirFS].
func Serve(dir string) http.Handler {
	return ServeFS(os.DirFS(dir))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
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

	// Set Cache-Control header
	if s.cacheControlFn != nil {
		cacheControl := s.cacheControlFn(r)
		if cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}
	}

	// Add 'Accept-Encoding' Vary header
	w.Header().Add("Vary", "Accept-Encoding")

	// Compressed (gzip)
	if acceptsGzip(r) {
		gzw := newGzipResponseWriter(w)
		defer gzw.Close()

		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(gzw, r, fileName, stat.ModTime(), content)
		return
	}

	http.ServeContent(w, r, fileName, stat.ModTime(), content)
}
