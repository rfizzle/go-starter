package server

import "go.uber.org/zap"

type Option func(*Server)

func WithAddress(address string) Option {
	return func(s *Server) {
		s.address = address
	}
}

func WithPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

func WithReadTimeout(readTimeout int) Option {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout int) Option {
	return func(s *Server) {
		s.writeTimeout = writeTimeout
	}
}

func WithIdleTimeout(idleTimeout int) Option {
	return func(s *Server) {
		s.idleTimeout = idleTimeout
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}
