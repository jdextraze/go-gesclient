package client

import (
	"fmt"
	"net"
)

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

func (a *ClientConnectionEventArgs) String() string {
	return fmt.Sprintf("&{remoteEndpoint:%s, connection:%s}", a.remoteEndpoint, a.connection.Name())
}

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

func (a *ClientReconnectingEventArgs) String() string {
	return fmt.Sprintf("&{connection:%s}", a.connection)
}

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

func (a *ClientClosedEventArgs) String() string {
	return fmt.Sprintf("&{reason:%s connection:%s}", a.reason, a.connection.Name())
}

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

func (a *ClientErrorEventArgs) String() string {
	return fmt.Sprintf("&{err:%s, connection:%s}", a.err.Error(), a.connection.Name())
}

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

func (a *ClientAuthenticationFailedEventArgs) String() string {
	return fmt.Sprintf("&{reason:%s, connection:%s}", a.reason, a.connection)
}
