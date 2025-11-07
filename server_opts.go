package srpc

import "github.com/tymbaca/srpc/logger"

type ServerOption func(s *Server)

func WithLogger(logger logger.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithStreamResponse controls response body writing process. Default: false.
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
func WithStreamResponse(val bool) ServerOption {
	return func(s *Server) {
		s.streamResponse = val
	}
}
