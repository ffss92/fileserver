package fileserver

import "net/http"

type CacheControlFunc func(r *http.Request) string

func noCacheAll(_ *http.Request) string {
	return "no-cache"
}
