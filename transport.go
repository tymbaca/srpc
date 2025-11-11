package srpc

import "github.com/tymbaca/srpc/transport"

var ErrListenerClosed = transport.ErrListenerClosed

var ErrListenerBadClose = transport.ErrListenerBadClose

// TODO: remove
type Dialer = transport.Dialer

type Listener = transport.Listener

type Conn = transport.Conn
