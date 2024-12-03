package fileserver

import (
	"errors"
	"io/fs"
	"net/http"
)

func ServeSPA(spa fs.FS, fallback string, opts ...ServerOptFn) http.Handler {
	h := New(spa, opts...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Path
		if target == "" {
			target = fallback
		}

		f, err := spa.Open(target)
		if err != nil {
			switch {
			case errors.Is(err, fs.ErrNotExist), errors.Is(err, fs.ErrInvalid):
				target = fallback
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			defer f.Close()

			stat, err := f.Stat()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if stat.IsDir() {
				target = fallback
			}
		}

		r.URL.Path = target
		h.ServeHTTP(w, r)
	})
}
