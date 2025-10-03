package srpc

import "github.com/tymbaca/srpc/logger"

type ServerOption func(s *Server)

func WithLogger(logger logger.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}
