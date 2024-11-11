package fileserver

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
)

var (
	// The path could not be found in the underlying [fs.FS]
	ErrFileNotFound = fmt.Errorf("file not found: %w", fs.ErrNotExist)
	// The underlying [fs.FS] returned a [fs.ErrInvalid] error. Check [fs.ValidPath] for path name rules.
	ErrInvalidPath = fmt.Errorf("invalid file path: %w", fs.ErrInvalid)
	// This server only supports GET requests. For any other method, the server's [ErrorHandlerFunc] is
	// called with this error.
	ErrInvalidMethod = errors.New("invalid http method")
)

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

// Creates a new file server using the [CalculateETag] to generate entity tags
// and creates an [os.DirFS] for dir.
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

	// Set Cache-Control header
	if s.cacheControlFn != nil {
		cacheControl := s.cacheControlFn(r)
		if cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}
	}

	// Append 'Accept-Encoding' to Vary header
	appendAcceptEncodingToVary(w)

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

// Appends Accept-Encoding to the currently set Vary header value.
func appendAcceptEncodingToVary(w http.ResponseWriter) {
	vary := w.Header().Get("Vary")
	if vary == "" {
		w.Header().Set("Vary", "Accept-Encoding")
	} else {
		w.Header().Set("Vary", vary+", Accept-Encoding")
	}
}
