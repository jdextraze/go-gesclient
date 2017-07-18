package internal

import (
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type message interface{}

type timerTickMessage struct{}

type startConnectionMessage struct {
	task               *tasks.CompletionSource
	endpointDiscoverer EndpointDiscoverer
}

func newStartConnectionMessage(
	resultChannel *tasks.CompletionSource,
	endpointDiscoverer EndpointDiscoverer,
) *startConnectionMessage {
	if resultChannel == nil {
		panic("resultChannel is nil")
	}
	if endpointDiscoverer == nil {
		panic("endpointDiscoverer is nil")
	}
	return &startConnectionMessage{
		task:               resultChannel,
		endpointDiscoverer: endpointDiscoverer,
	}
}

func (m *startConnectionMessage) Task() *tasks.CompletionSource          { return m.task }
func (m *startConnectionMessage) EndpointDiscoverer() EndpointDiscoverer { return m.endpointDiscoverer }

type closeConnectionMessage struct {
	reason string
	error  error
}

func newCloseConnectionMessage(
	reason string,
	error error,
) *closeConnectionMessage {
	return &closeConnectionMessage{
		reason: reason,
		error:  error,
	}
}

func (m *closeConnectionMessage) Reason() string { return m.reason }
func (m *closeConnectionMessage) Error() error   { return m.error }

type establishTcpConnectionMessage struct {
	endpoints *NodeEndpoints
}

func newEstablishTcpConnectionMessage(endpoints *NodeEndpoints) *establishTcpConnectionMessage {
	return &establishTcpConnectionMessage{
		endpoints: endpoints,
	}
}

func (m *establishTcpConnectionMessage) Endpoints() *NodeEndpoints { return m.endpoints }

type tcpConnectionEstablishedMessage struct {
	connection *client.PackageConnection
}

func newTcpConnectionEstablishedMessage(connection *client.PackageConnection) *tcpConnectionEstablishedMessage {
	if connection == nil {
		panic("connection is nil")
	}
	return &tcpConnectionEstablishedMessage{
		connection: connection,
	}
}

func (m *tcpConnectionEstablishedMessage) Connection() *client.PackageConnection { return m.connection }

type tcpConnectionClosedMessage struct {
	connection  *client.PackageConnection
	socketError error
}

func newTcpConnectionClosedMessage(
	connection *client.PackageConnection,
	socketError error,
) *tcpConnectionClosedMessage {
	if connection == nil {
		panic("connection is nil")
	}
	return &tcpConnectionClosedMessage{
		connection:  connection,
		socketError: socketError,
	}
}

func (m *tcpConnectionClosedMessage) Connection() *client.PackageConnection { return m.connection }
func (m *tcpConnectionClosedMessage) SocketError() error                    { return m.socketError }

type startOperationMessage struct {
	operation  client.Operation
	maxRetries int
	timeout    time.Duration
}

type startSubscriptionMessage struct {
	source              *tasks.CompletionSource
	streamId            string
	resolveLinkTos      bool
	userCredentials     *client.UserCredentials
	eventAppeared       client.EventAppearedHandler
	subscriptionDropped client.SubscriptionDroppedHandler
	maxRetries          int
	timeout             time.Duration
}

type startPersistentSubscriptionMessage struct {
	source              *tasks.CompletionSource
	subscriptionId      string
	streamId            string
	bufferSize          int
	userCredentials     *client.UserCredentials
	eventAppeared       client.EventAppearedHandler
	subscriptionDropped client.SubscriptionDroppedHandler
	maxRetries          int
	timeout             time.Duration
}

type handleTcpPackageMessage struct {
	connection *client.PackageConnection
	pkg        *client.Package
}

type tcpConnectionErrorMessage struct {
	connection *client.PackageConnection
	error      error
}
