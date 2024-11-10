package fileserver

import (
	"errors"
	"net/http"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

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
