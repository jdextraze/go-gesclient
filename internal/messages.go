package internal

import (
	"time"
	"github.com/jdextraze/go-gesclient/models"
)

type message interface{}

type timerTickMessage struct{}

type startConnectionMessage struct {
	resultChannel      chan interface{}
	endpointDiscoverer EndpointDiscoverer
}

func newStartConnectionMessage(
	resultChannel chan interface{},
	endpointDiscoverer EndpointDiscoverer,
) *startConnectionMessage {
	if resultChannel == nil {
		panic("resultChannel is nil")
	}
	if endpointDiscoverer == nil {
		panic("endpointDiscoverer is nil")
	}
	return &startConnectionMessage{
		resultChannel:      resultChannel,
		endpointDiscoverer: endpointDiscoverer,
	}
}

func (m *startConnectionMessage) ResultChannel() chan interface{}        { return m.resultChannel }
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
	connection *PackageConnection
}

func newTcpConnectionEstablishedMessage(connection *PackageConnection) *tcpConnectionEstablishedMessage {
	if connection == nil {
		panic("connection is nil")
	}
	return &tcpConnectionEstablishedMessage{
		connection: connection,
	}
}

func (m *tcpConnectionEstablishedMessage) Connection() *PackageConnection { return m.connection }

type tcpConnectionClosedMessage struct {
	connection  *PackageConnection
	socketError error
}

func newTcpConnectionClosedMessage(
	connection *PackageConnection,
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

func (m *tcpConnectionClosedMessage) Connection() *PackageConnection { return m.connection }
func (m *tcpConnectionClosedMessage) SocketError() error             { return m.socketError }

// TODO complete messages
type startOperationMessage struct {
	operation models.Operation
	maxRetries int
	timeout time.Duration
}

type startSubscriptionMessage struct {
	source <-chan models.Subscription
	streamId string
	resolveLinkTos bool
	userCredentials *models.UserCredentials
	eventAppeared func(s models.Subscription, r models.ResolvedEvent)
	subscriptionDropped func(s models.Subscription, dr models.SubscriptionDropReason, err error)
	maxRetries int
	timeout time.Duration
}

type startPersistentSubscriptionMessage struct {}

type handleTcpPackageMessage struct {
	connection *PackageConnection
	pkg *models.Package
}

type tcpConnectionErrorMessage struct {
	connection  *PackageConnection
	error error
}
