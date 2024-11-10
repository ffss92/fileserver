package main

import (
	"flag"
	"net/http"

	"github.com/ffss92/fileserver"
)

type config struct {
	addr string
	dir  string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8000", "sets the server addr")
	flag.StringVar(&cfg.dir, "dir", "", "sets the directory to be served")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", fileserver.Serve("testdata")))
	http.ListenAndServe(":4000", mux)
}
