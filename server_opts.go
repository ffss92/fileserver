package fileserver

type ServerOptFn func(s *Server)

// Adds a custom ETag function to the server.
func WithETagFunc(etagFn ETagFunc) ServerOptFn {
	return func(s *Server) {
		s.etagFn = etagFn
	}
}

// Adds a custom error handler function to the server that's called
// whenever an error happens.
//
// By default, an 404 response is sent for [ErrFileNotFound], a 405 response is sent to [ErrInvalidMethod]
// and a 400 response is sent to [ErrInvalidPath]. For unknown errors, the server responds with a 500 Internal Server Error response.
func WithErrorHandler(errHandler ErrorHandlerFunc) ServerOptFn {
	return func(s *Server) {
		s.errHandler = errHandler
	}
}

// Adds a custom Cache-Control header function. The default behavior is to set the Cache-Control header
// value to "no-cache", which is equivalent to "public, max-age=0, must-revalidate". If a nil function
// is provided, the server won't set any value to the Cache-Control header.
func WithCacheControlFunc(cacheControlFn CacheControlFunc) ServerOptFn {
	return func(s *Server) {
		s.cacheControlFn = cacheControlFn
	}
}
