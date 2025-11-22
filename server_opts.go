package srpc

import "github.com/tymbaca/srpc/logger"

// ServerOption modifies the server when passed into [NewServer].
type ServerOption func(s *Server)

// WithLogger sets provided logger in the server.
func WithLogger(logger logger.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithConnErrorHandler sets handler for errors happening when handling
// connections. If handler itself returns an error (even if it's the original error)
// server will log it, using it's logger. If handler returns nil, server won't
// original error.
func WithConnErrorHandler(handler func(error) error) ServerOption {
	return func(s *Server) {
		s.connErrorHandler = handler
	}
}

// WithStreamingResponse controls response body writing process. Default: false.
//
// If val is false, then service return values will be fully encoded into bytes buffer before
// sending to the client.
//
// If val is true, then values will be encoded and send in a
// streaming manner as they being sent, optimizing memory usage in case of big
// payloads.
//
// Note: with this option set to true client won't be able to get descriptive
// error message if server gets error from [Codec.Decode], because that error
// will be produced when sending process has already been started.
func WithStreamingResponse(val bool) ServerOption {
	return func(s *Server) {
		s.streamResponse = val
	}
}
