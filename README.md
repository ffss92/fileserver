# fileserver

Enhanced file server for Go.

1. Provides ETag header generation (hex encoded md5 hash);
1. Compression with `gzip`.

For the full documentation, [click here](https://pkg.go.dev/github.com/ffss92/fileserver).

## Installation

To get started, run:

```bash
go get github.com/ffss92/fileserver
```

## Usage

To get started using `fileserver` as a package in your application, simply mount it to your current router:

```go
static := os.DirFS("static")
mux := http.NewServeMux()
// Stripping prefix is important here, or else your files won't be found.
mux.Handle("/static/", http.StripPrefix("/static/", fileserver.ServeFS(static)))
// Or
mux.Handle("/static/", http.StripPrefix("/static/", fileserver.Serve("assets")))
```

## Roadmap

- Attempt to serve `index.html` instead of returning a 404 if a
  directory is requested. For example, requests to `/static/dir`
  will attempt serve `/static/dir/index.html`, if present.
