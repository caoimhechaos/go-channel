package channel

import (
	"errors"
	"net/url"
	"time"
)

// Function creating a channel from the given URL.
type URLHandler func(*url.URL, time.Duration) (Channel, error)

// The list of all regsitered URL handlers.
var urlHandlers map[string]URLHandler = make(map[string]URLHandler)

// Register a given handler for the specified URL pattern.
func RegisterURLHandler(schema string, uh URLHandler) {
	urlHandlers[schema] = uh
}

// Connect to the destintation specified in "dest" (which should be an URL).
func ChannelFromString(dest string, timeout time.Duration) (
	c Channel, err error) {
	var u *url.URL
	u, err = url.Parse(dest)
	if err != nil {
		var e2 error
		// Fall back to thinking it might be a TCP host:port pair.
		c, e2 = NewSocketChannel("tcp", dest, timeout)
		if e2 != nil {
			c, e2 = NewSocketChannel("udp", dest, timeout)
			if e2 != nil {
				return
			}
		}
	}
	return ChannelFromURL(u, timeout)
}

// Connect to the destination specified in "dest".
// A handler for the URL schema must be registered.
func ChannelFromURL(u *url.URL, timeout time.Duration) (
	c Channel, err error) {
	var h URLHandler
	var b bool

	h, b = urlHandlers[u.Scheme]
	if b {
		return h(u, timeout)
	}

	return nil, errors.New("No handler registered for schema " + u.Scheme)
}
