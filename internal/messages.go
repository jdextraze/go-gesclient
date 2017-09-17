package internal

import (
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type message interface {
	MessageID() int
}

type timerTickMessage struct{}

func (m *timerTickMessage) MessageID() int { return 0 }

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
func (m *startConnectionMessage) MessageID() int                         { return 1 }

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
func (m *closeConnectionMessage) MessageID() int { return 2 }

type establishTcpConnectionMessage struct {
	endpoints *NodeEndpoints
}

func newEstablishTcpConnectionMessage(endpoints *NodeEndpoints) *establishTcpConnectionMessage {
	return &establishTcpConnectionMessage{
		endpoints: endpoints,
	}
}

func (m *establishTcpConnectionMessage) Endpoints() *NodeEndpoints { return m.endpoints }
func (m *establishTcpConnectionMessage) MessageID() int            { return 3 }

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
func (m *tcpConnectionEstablishedMessage) MessageID() int                        { return 4 }

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
func (m *tcpConnectionClosedMessage) MessageID() int                        { return 5 }

type startOperationMessage struct {
	operation  client.Operation
	maxRetries int
	timeout    time.Duration
}

func newStartOperationMessage(
	operation client.Operation,
	maxRetries int,
	timeout time.Duration,
) *startOperationMessage {
	return &startOperationMessage{
		operation:  operation,
		maxRetries: maxRetries,
		timeout:    timeout,
	}
}

func (m *startOperationMessage) MessageID() int { return 6 }

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

func newStartSubscriptionMessage(
	source *tasks.CompletionSource,
	streamId string,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
	maxRetries int,
	timeout time.Duration,
) *startSubscriptionMessage {
	return &startSubscriptionMessage{
		source:              source,
		streamId:            streamId,
		resolveLinkTos:      resolveLinkTos,
		userCredentials:     userCredentials,
		eventAppeared:       eventAppeared,
		subscriptionDropped: subscriptionDropped,
		maxRetries:          maxRetries,
		timeout:             timeout,
	}
}

func (m *startSubscriptionMessage) MessageID() int { return 7 }

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

func (m *startPersistentSubscriptionMessage) MessageID() int { return 8 }

func NewStartPersistentSubscriptionMessage(
	source *tasks.CompletionSource,
	subscriptionId string,
	streamId string,
	bufferSize int,
	userCredentials *client.UserCredentials,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
	maxRetries int,
	timeout time.Duration,
) *startPersistentSubscriptionMessage {
	return &startPersistentSubscriptionMessage{
		source:              source,
		subscriptionId:      subscriptionId,
		streamId:            streamId,
		bufferSize:          bufferSize,
		userCredentials:     userCredentials,
		eventAppeared:       eventAppeared,
		subscriptionDropped: subscriptionDropped,
		maxRetries:          maxRetries,
		timeout:             timeout,
	}
}

type handleTcpPackageMessage struct {
	connection *client.PackageConnection
	pkg        *client.Package
}

func newHandleTcpPackageMessage(
	connection *client.PackageConnection,
	pkg *client.Package,
) *handleTcpPackageMessage {
	if connection == nil {
		panic("connection is nil")
	}
	if pkg == nil {
		panic("connection is nil")
	}
	return &handleTcpPackageMessage{
		connection: connection,
		pkg:        pkg,
	}
}

func (m *handleTcpPackageMessage) MessageID() int { return 9 }

type tcpConnectionErrorMessage struct {
	connection *client.PackageConnection
	error      error
}

func newTcpConnectionErrorMessage(
	connection *client.PackageConnection,
	err error,
) *tcpConnectionErrorMessage {
	if connection == nil {
		panic("connection is nil")
	}
	if err == nil {
		panic("error is nil")
	}
	return &tcpConnectionErrorMessage{
		connection: connection,
		error:      err,
	}
}

func (m *tcpConnectionErrorMessage) MessageID() int { return 10 }
