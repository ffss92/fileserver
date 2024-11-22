package fileserver

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
)

var (
	// The path could not be found in the underlying [fs.FS]
	ErrFileNotFound = fmt.Errorf("file not found: %w", fs.ErrNotExist)
	// The underlying [fs.FS] returned a [fs.ErrInvalid] error. Check [fs.ValidPath] for path name rules.
	ErrInvalidPath = fmt.Errorf("invalid file path: %w", fs.ErrInvalid)
	// This server only supports GET and HEAD requests. For any other method, the server's [ErrorHandlerFunc] is
	// called with this error.
	ErrInvalidMethod = errors.New("invalid http method")
)

type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrFileNotFound):
		http.Error(w, "file not found", http.StatusNotFound)
	case errors.Is(err, ErrInvalidPath):
		http.Error(w, "invalid file path", http.StatusBadRequest)
	case errors.Is(err, ErrInvalidMethod):
		http.Error(w, "only GET is supported", http.StatusMethodNotAllowed)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
