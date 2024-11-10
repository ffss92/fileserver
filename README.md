# fileserver

Simple file server written in Go.

1. Provides ETag header generation (hex encoded md5 hash);
2. Compression with `gzip`.

## Usage

### 1. Package

To get started using `fileserver` as a package in your application, simply mount it to your current router:

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

If you just want to spin a local fileserver quickly, just run:

```bash
./fileserver -dir assets
```
