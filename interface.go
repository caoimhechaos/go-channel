package channel

import (
	"net"
	"time"
)

/*
Channels are essentially read/write/close interfaces with an additional
option to force the channel to have at least one backend, and to switch
to a different backend (which may or may not do something).
*/
type Channel interface {
	net.Conn

	// Delay until either we have at least 1 backend or until the deadline
	// has expired. In case the deadline expires, an error is returned.
	WaitForNonEmpty(deadline time.Duration) error

	// Suggest to the channel to pick the next backend. This can be used
	// for load balancing or other reasons. There's no guarantee that the
	// backend will actually change though.
	NextBackend()

	// Total number of backends currently connected.
	NumBackends() uint64

	// List of all channels contained in this one. Can be useful to e.g.
	// perform broadcast RPCs.
	GetAllSubchannels() []Channel
}
