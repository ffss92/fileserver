package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ffss92/fileserver"
)

type config struct {
	addr   string
	dir    string
	silent bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.dir, "dir", "", "sets the target directory")
	flag.StringVar(&cfg.addr, "addr", ":8000", "sets the server addr")
	flag.BoolVar(&cfg.silent, "silent", false, "disables request logging")
	flag.Parse()

	if cfg.dir == "" {
		log.Fatal("please provide a target")
	}
	if cfg.silent {
		log.SetOutput(io.Discard)
	}

	h := http.StripPrefix("/", fileserver.Serve(cfg.dir))

	log.Printf("Serving %q on %q\n", cfg.dir, cfg.addr)
	log.Fatal(http.ListenAndServe(cfg.addr, logger(h)))
}

type loggerResponseWriter struct {
	status int
	http.ResponseWriter
}

func (l *loggerResponseWriter) WriteHeader(status int) {
	l.status = status
	l.ResponseWriter.WriteHeader(status)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		lw := &loggerResponseWriter{status: http.StatusOK, ResponseWriter: w}
		defer func() {
			log.Printf("%s %s - %d (%s)", r.Method, r.URL.RequestURI(), lw.status, time.Since(t))
		}()
		next.ServeHTTP(lw, r)
	})
}
