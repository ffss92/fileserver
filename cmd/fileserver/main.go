package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ffss92/fileserver"
)

type config struct {
	addr     string
	spa      bool
	fallback string
	silent   bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8000", "Sets the server listen address.")
	flag.BoolVar(&cfg.spa, "spa", false, "Sets the server in SPA mode.")
	flag.StringVar(&cfg.fallback, "fallback", "index.html", "Sets the SPA fallback file.")
	flag.BoolVar(&cfg.silent, "silent", false, "Disables request logging.")
	flag.Parse()

	dir := flag.Arg(0)
	if dir == "" {
		log.Fatal("please provide a target")
	}

	if cfg.silent {
		log.SetOutput(io.Discard)
	}

	var h http.Handler
	if cfg.spa {
		log.Printf("Serving %q on %q in SPA mode\n", dir, cfg.addr)
		h = http.StripPrefix("/", fileserver.ServeSPA(os.DirFS(dir), cfg.fallback))
	} else {
		log.Printf("Serving %q on %q\n", dir, cfg.addr)
		h = http.StripPrefix("/", fileserver.Serve(dir))
	}

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
