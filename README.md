# fileserver

Simple file server written in Go.

1. Provides ETag header generation (hex encoded md5 hash);
2. Compression with `gzip`.

This package is still under development. It currently always set `Cache-Control`
to `no-cache` (`public,max-age=0,must-revalidate`).

## Usage

### 1. Package

To get started using `fileserver` as a package in your application, simply mount it to your current router:

```bash
go get github.com/ffss92/fileserver
```

```go
mux := http.NewServeMux()
// Stripping prefix is important here, or else your files won't be found.
mux.Handle("/static/", http.StripPrefix("/static/", fileserver.Serve("assets")))
```

You can add custom configuration by calling `fileserver.New`.

```go
assets := os.DirFS("web/assets")
fileServer := fileserver.New(assets)
```

### 2. CLI

First, install the server by running:

```bash
go install github.com/ffss92/fileserver/cmd/fileserver@latest
```

If you just want to spin a local fileserver quickly, just run:

```bash
./fileserver -dir assets
```

You can check all available flags running:

```bash
./fileserver -h
```
