package client

import "net"

type ClientConnectionEventArgs struct {
	remoteEndpoint net.Addr
	connection     Connection
}

func NewClientConnectionEventArgs(
	remoteEndpoint net.Addr,
	connection Connection,

) *ClientConnectionEventArgs {
	return &ClientConnectionEventArgs{
		remoteEndpoint: remoteEndpoint,
		connection:     connection,
	}
}

func (a *ClientConnectionEventArgs) RemoteEndpoint() net.Addr { return a.remoteEndpoint }

func (a *ClientConnectionEventArgs) Connection() Connection { return a.connection }

//

type ClientReconnectingEventArgs struct {
	connection Connection
}

func NewClientReconnectingEventArgs(
	connection Connection,

) *ClientReconnectingEventArgs {
	return &ClientReconnectingEventArgs{
		connection: connection,
	}
}

func (a *ClientReconnectingEventArgs) Connection() Connection { return a.connection }

//

type ClientClosedEventArgs struct {
	reason     string
	connection Connection
}

func NewClientClosedEventArgs(
	reason string,
	connection Connection,
) *ClientClosedEventArgs {
	return &ClientClosedEventArgs{
		reason:     reason,
		connection: connection,
	}
}

func (a *ClientClosedEventArgs) Reason() string { return a.reason }

func (a *ClientClosedEventArgs) Connection() Connection { return a.connection }

//

type ClientErrorEventArgs struct {
	err        error
	connection Connection
}

func NewClientErrorEventArgs(
	err error,
	connection Connection,
) *ClientErrorEventArgs {
	return &ClientErrorEventArgs{
		err:        err,
		connection: connection,
	}
}

func (a *ClientErrorEventArgs) Error() error { return a.err }

func (a *ClientErrorEventArgs) Connection() Connection { return a.connection }

//

type ClientAuthenticationFailedEventArgs struct {
	reason     string
	connection Connection
}

func NewClientAuthenticationFailedEventArgs(
	reason string,
	connection Connection,
) *ClientAuthenticationFailedEventArgs {
	return &ClientAuthenticationFailedEventArgs{
		reason:     reason,
		connection: connection,
	}
}

func (a *ClientAuthenticationFailedEventArgs) Reason() string { return a.reason }

func (a *ClientAuthenticationFailedEventArgs) Connection() Connection { return a.connection }
