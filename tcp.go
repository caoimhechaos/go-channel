package channel

import (
	"io"
	"net"
	"net/url"
	"time"
)

func init() {
	RegisterURLHandler("tcp", NewTCPChannel)
	RegisterURLHandler("udp", NewUDPChannel)
}

// A TCP channel is a channel over a simple TCP connection with just one peer.
type SocketChannel struct {
	peerAddr      string
	conn          net.Conn
	timeout       time.Duration
	proto         string
	manualTimeout bool
}

// Create a new TCP channel for connecting to "dest".
// If "timeout" is greater than 0, set deadlines on all operations to
// the current time plus "timeout".
func NewTCPChannel(dest *url.URL, timeout time.Duration) (ch Channel, err error) {
	return NewSocketChannel("tcp", dest.Host, timeout)
}

// Create a new UDP channel for connecting to "dest".
// If "timeout" is greater than 0, set deadlines on all operations to
// the current time plus "timeout".
func NewUDPChannel(dest *url.URL, timeout time.Duration) (ch Channel, err error) {
	return NewSocketChannel("udp", dest.Host, timeout)
}

// Encapsulate the specified socket in a channel.
func NewChannelFromSocket(connection net.Conn, timeout time.Duration) Channel {
	return &SocketChannel{
		peerAddr:      connection.RemoteAddr().String(),
		conn:          connection,
		timeout:       timeout,
		proto:         connection.RemoteAddr().Network(),
		manualTimeout: false,
	}
}

// Create a new socket channel for connecting to "dest".
// If "timeout" is greater than 0, set deadlines on all operations to
// the current time plus "timeout".
func NewSocketChannel(proto, dest string, timeout time.Duration) (ch Channel, err error) {
	var conn net.Conn

	if timeout > 0 {
		conn, err = net.DialTimeout(proto, dest, timeout)
	} else {
		conn, err = net.Dial(proto, dest)
	}
	if err != nil {
		return
	}

	return &SocketChannel{
		peerAddr:      dest,
		conn:          conn,
		timeout:       timeout,
		proto:         proto,
		manualTimeout: false,
	}, nil
}

// Socket channels are connected as soon as the constructor returns.
func (*SocketChannel) WaitForNonEmpty(unused time.Duration) error {
	return nil
}

// On single-destination Socket connections, switching channels is a no-op.
func (*SocketChannel) NextBackend() {
}

// Closing a Socket channel closes the underlying Socket socket.
func (t *SocketChannel) Close() error {
	return t.conn.Close()
}

// Local address is whatever the connection feels like.
func (t *SocketChannel) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

// Remote address is whatever the connection thinks it is.
func (t *SocketChannel) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

// Set the deadline for reading and writing and stop automated
// deadline management.
func (s *SocketChannel) SetDeadline(t time.Time) error {
	s.manualTimeout = true
	return s.conn.SetDeadline(t)
}

// Set the deadline for reading and stop automated deadline management.
func (s *SocketChannel) SetReadDeadline(t time.Time) error {
	s.manualTimeout = true
	return s.conn.SetReadDeadline(t)
}

// Set the deadline for writing and stop automated deadline management.
func (s *SocketChannel) SetWriteDeadline(t time.Time) error {
	s.manualTimeout = true
	return s.conn.SetWriteDeadline(t)
}

// Reading from a Socket channel is basically just reading from the
// underlying Socket connection.
func (t *SocketChannel) Read(p []byte) (n int, err error) {
	if t.timeout > 0 {
		var deadline time.Time = time.Now().Add(t.timeout)
		t.conn.SetReadDeadline(deadline)
	}
	return t.conn.Read(p)
}

// Sockets only have a single backend, so the largest subset of
// channels making up this channel is this channel itself.
func (t *SocketChannel) GetAllSubchannels() []Channel {
	return []Channel{t}
}

// Sockets only have a single backend.
func (s *SocketChannel) NumBackends() uint64 {
	return 1
}

// Writing to a Socket channel is slightly more involved than reading from
// it. If the channel was closed, it must be reestablished.
func (t *SocketChannel) Write(p []byte) (n int, err error) {
	var deadline time.Time
	if t.timeout > 0 {
		deadline = time.Now().Add(t.timeout)
	}
	for {
		if deadline.UnixNano() > 0 {
			t.conn.SetWriteDeadline(deadline)
		}
		n, err = t.conn.Write(p)
		if err == nil {
			return
		} else {
			if err == io.ErrClosedPipe {
				t.conn.Close()
				// Don't try to wait beyond the deadline, that's pointless.
				if time.Now().After(deadline) {
					return
				}
				if t.timeout > 0 {
					t.conn, err = net.DialTimeout(t.proto, t.peerAddr, t.timeout)
				} else {
					t.conn, err = net.Dial(t.proto, t.peerAddr)
				}
				if err != nil {
					return
				}
			} else {
				return
			}
		}
	}
}
