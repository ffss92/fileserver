package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ffss92/fileserver"
)

type config struct {
	addr string
	dir  string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.dir, "dir", "", "sets the target directory")
	flag.StringVar(&cfg.addr, "addr", ":8000", "sets the server addr")
	flag.Parse()

	if cfg.dir == "" {
		log.Fatal("please provide a target")
	}

	log.Printf("serving %q on %q\n", cfg.dir, cfg.addr)
	log.Fatal(http.ListenAndServe(cfg.addr, http.StripPrefix("/", fileserver.Serve(cfg.dir))))
}
