package fileserver

type ServerOptFn func(s *Server)

func WithETagFunc(etagFn ETagFunc) ServerOptFn {
	return func(s *Server) {
		s.etagFn = etagFn
	}
}

func WithErrorHandler(errHandler ErrorHandler) ServerOptFn {
	return func(s *Server) {
		s.errHandler = errHandler
	}
}
