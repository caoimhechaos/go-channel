package channel

/*
The Go Channel project supplies channels for performing atomic read/write
operations on a collection of interchangable endpoints.
The semantics of a channel are:

 * A channel connects to one or more endpoints.
 * It's possible to wait for a channel to have at least 1 backend.
 * Send operations must send all data at the same time.
 * If an error occurs during sending and the data has not been sent
   yet, the send operation is retried on a different backend.
 * Otherwise, a write error is returned and subsequent sends will
   go to a different backend.
 * Switching backends can, in general, be forced (if only one backend
   is connected, it will be a no-op).
 * Channels may choose to switch backends on each send operation depending
   on the channel type.
 * If a read from a channel fails, an error will always be reported.
*/
