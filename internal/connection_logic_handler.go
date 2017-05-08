package internal

import (
	"github.com/jdextraze/go-gesclient/models"
	"net"
	"time"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"sync/atomic"
)

type ConnectionLogicHandler interface {
	TotalOperationCount() int
	EnqueueMessage(msg message) error
	Connected() models.EventHandlers
	Disconnected() models.EventHandlers
	Reconnecting() models.EventHandlers
	Closed() models.EventHandlers
	ErrorOccurred() models.EventHandlers
	AuthenticationFailed() models.EventHandlers
}

type heartbeatInfo struct {
	LastPackageNumber int
	IsIntervalStage   bool
	Timestamp         time.Duration
}

type reconnectionInfo struct {
	ReconnectionAttempt int
	Timestamp           time.Duration
}

type authInfo struct {
	CorrelationId uuid.UUID
	Timestamp     time.Duration
}

type connectionState int

const (
	connectionState_Init       connectionState = iota
	connectionState_Connecting
	connectionState_Connected
	connectionState_Closed
)

type connectingPhase int

const (
	connectingPhase_Invalid                connectingPhase = iota
	connectingPhase_Reconnecting
	connectingPhase_EndpointDiscovery
	connectingPhase_ConnectionEstablishing
	connectingPhase_Authentication
	connectingPhase_Connected
)

type connectionLogicHandler struct {
	connected             *eventHandlers
	disconnected          *eventHandlers
	reconnecting          *eventHandlers
	closed                *eventHandlers
	errorOccurred         *eventHandlers
	authenticationFailed  *eventHandlers
	esConnection          models.Connection
	settings              *models.ConnectionSettings
	queue                 *simpleQueuedHandler
	timer                 *time.Timer
	endpointDiscoverer    EndpointDiscoverer
	startTime             time.Time
	reconInfo             reconnectionInfo
	heartbeatInfo         heartbeatInfo
	authInfo              authInfo
	lastTimeoutsTimestamp time.Duration
	operations            OperationManager
	subscriptions         SubscriptionManager
	state                 connectionState
	connectingPhase       connectingPhase
	wasConnected          int32
	packageNumber         int
	connection            *PackageConnection
}

func newConnectionLogicHandler(
	connection models.Connection,
	settings *models.ConnectionSettings,
) (*connectionLogicHandler, error) {
	if connection == nil {
		return nil, errors.New("connection is nil")
	}
	if settings == nil {
		return nil, errors.New("settings is nil")
	}

	queue := newSimpleQueuedHandler(settings.MaxQueueSize())

	obj := &connectionLogicHandler{
		connected:            newEventHandlers(),
		disconnected:         newEventHandlers(),
		reconnecting:         newEventHandlers(),
		closed:               newEventHandlers(),
		errorOccurred:        newEventHandlers(),
		authenticationFailed: newEventHandlers(),
		esConnection:         connection,
		settings:             settings,
		queue:                queue,
		startTime:            time.Now(),
	}

	queue.RegisterHandler(startConnectionMessage{}, obj.startConnection)
	queue.RegisterHandler(closeConnectionMessage{}, obj.closeConnection)

	queue.RegisterHandler(startOperationMessage{}, obj.startOperation)
	queue.RegisterHandler(startSubscriptionMessage{}, obj.startSubscription)
	queue.RegisterHandler(startPersistentSubscriptionMessage{}, obj.startPersistentSubscription)

	queue.RegisterHandler(establishTcpConnectionMessage{}, obj.establishTcpConnection)
	queue.RegisterHandler(tcpConnectionEstablishedMessage{}, obj.tcpConnectionEstablished)
	queue.RegisterHandler(tcpConnectionErrorMessage{}, obj.tcpConnectionError)
	queue.RegisterHandler(tcpConnectionClosedMessage{}, obj.tcpConnectionClosed)
	queue.RegisterHandler(handleTcpPackageMessage{}, obj.handleTcpPackage)

	queue.RegisterHandler(timerTickMessage{}, obj.timerTick)

	obj.timer = time.AfterFunc(models.TimerPeriod, func() { obj.EnqueueMessage(timerTickMessage{}) })

	return obj, nil
}

func (h *connectionLogicHandler) TotalOperationCount() int {
	return h.operations.TotalOperationCount()
}

func (h *connectionLogicHandler) EnqueueMessage(msg message) error {
	_, isTimerTickMessage := msg.(timerTickMessage)
	if h.settings.VerboseLogging() && !isTimerTickMessage {
		log.Debug("enqueuing message %v", msg)
	}
	return h.queue.EnqueueMessage(msg)
}

func (h *connectionLogicHandler) startConnection(msg message) error {
	startConnectionMessage, isStartConnectionMessage := msg.(startConnectionMessage)
	if !isStartConnectionMessage {
		panic("invalid message type. expected startConnectionMessage")
	}
	if startConnectionMessage.resultChannel == nil {
		panic("startConnectionMessage.resultChannel is nil")
	}
	if startConnectionMessage.endpointDiscoverer == nil {
		panic("startConnectionMessage.endpointDiscoverer is nil")
	}
	log.Debug("Start connection")
	switch h.state {
	case connectionState_Init:
		h.endpointDiscoverer = startConnectionMessage.endpointDiscoverer
		h.state = connectionState_Connecting
		h.connectingPhase = connectingPhase_Reconnecting
		h.discoverEndpoint(startConnectionMessage.resultChannel)
	case connectionState_Connecting, connectionState_Connected:
		startConnectionMessage.resultChannel <- fmt.Errorf(
			"EventStoreConnection '%s' is already active", h.esConnection.Name())
	case connectionState_Closed:
		startConnectionMessage.resultChannel <- fmt.Errorf(
			"EventStoreConnection '%s' is closed", h.esConnection.Name())
	default:
		startConnectionMessage.resultChannel <- fmt.Errorf("Unknown state '%v'", h.state)
	}
	return nil
}

func (h *connectionLogicHandler) discoverEndpoint(ch chan interface{}) {
	log.Debug("Discover endpoint")

	if h.state != connectionState_Connecting {
		return
	}
	if h.connectingPhase != connectingPhase_Reconnecting {
		return
	}

	h.connectingPhase = connectingPhase_EndpointDiscovery

	go func() {
		discovered := <-h.endpointDiscoverer.DiscoverAsync(h.connection.RemoteEndpoint())
		if discovered.error != nil {
			h.EnqueueMessage(newCloseConnectionMessage(
				"Failed to resolve TCP end point to which to connect.",
				discovered.error,
			))
			if ch != nil {
				ch <- errors.New("Cannot resolve target endpoint")
			}
			return
		}
		h.EnqueueMessage(newEstablishTcpConnectionMessage(discovered.nodeEndpoints))
		if ch != nil {
			ch <- nil
		}
	}()
}

func (h *connectionLogicHandler) closeConnection(msg message) error {
	m := msg.(*closeConnectionMessage)

	if h.state == connectionState_Closed {
		log.Debugf("CloseConnection IGNORED because is ESConnection is CLOSED, reason %s, exception %v.",
			m.reason, m.error)
		return nil
	}

	log.Debugf("CloseConnection, reason %s, exception %v.", m.reason, m.error)

	h.state = connectionState_Closed

	h.timer.Stop()
	h.timer = nil
	h.operations.CleanUp()
	h.subscriptions.CleanUp()
	h.closeTcpConnection(m.reason)

	log.Infof("Closed. Reason: %s", m.reason)

	if m.error != nil {
		h.raiseErrorOccurred(m.error)
	}

	h.raiseClosed(m.reason)

	return nil
}

func (h *connectionLogicHandler) closeTcpConnection(reason string) {
	if h.connection == nil {
		log.Debug("CloseTcpConnection IGNORED because _connection == null")
		return
	}

	log.Debug("CloseTcpConnection")
	h.connection.Close(reason)
	h.tcpConnectionClosed(&tcpConnectionClosedMessage{h.connection, nil})
	h.connection = nil
}

func (h *connectionLogicHandler) startOperation(msg message) error {
	m := msg.(*startOperationMessage)

	switch h.state {
	case connectionState_Init:
		m.operation.Fail(fmt.Errorf("EventStoreConnection '%s' is not active", h.esConnection.Name()))
	case connectionState_Connecting:
		log.Debugf("StartOperation enqueue %s, %d, %v", m.operation, m.maxRetries, m.timeout)
		h.operations.EnqueueOperation()
	}

	return nil
}

func (h *connectionLogicHandler) startSubscription(msg message) error {
	return nil
}

func (h *connectionLogicHandler) startPersistentSubscription(msg message) error {
	return nil
}

func (h *connectionLogicHandler) establishTcpConnection(msg message) error {
	establishTcpConnection, isEstablishTcpConnection := msg.(establishTcpConnectionMessage)
	if !isEstablishTcpConnection {
		panic("invalid message type. expected establishTcpConnection")
	}
	var tcpEndpoint *net.TCPAddr
	if h.settings.UseSslConnection() {
		if establishTcpConnection.endpoints.secureTcpEndpoint == nil {
			tcpEndpoint = establishTcpConnection.endpoints.tcpEndpoint
		} else {
			tcpEndpoint = establishTcpConnection.endpoints.secureTcpEndpoint
		}
	} else {
		tcpEndpoint = establishTcpConnection.endpoints.tcpEndpoint
	}
	if tcpEndpoint == nil {
		h.closeConnection(closeConnectionMessage{
			reason: "No endpoint to node specified.",
		})
	}
	if h.state != connectionState_Connecting {
		return nil
	}
	if h.connectingPhase != connectingPhase_EndpointDiscovery {
		return nil
	}
	h.connectingPhase = connectingPhase_ConnectionEstablishing
	h.connection = newPackageConnection(log, tcpEndpoint, uuid.NewV4(), h.settings.UseSslConnection(),
		h.settings.TargetHost(), h.settings.ValidateService(), h.settings.ClientConnectionTimeout(),
		func(c *PackageConnection, p *models.Package) { h.EnqueueMessage(&handleTcpPackageMessage{c, p}) },
		func(c *PackageConnection, err error) { h.EnqueueMessage(&tcpConnectionErrorMessage{c, err}) },
		func(c *PackageConnection) { h.EnqueueMessage(&tcpConnectionEstablishedMessage{c}) },
		func(c *PackageConnection, err error) { h.EnqueueMessage(&tcpConnectionClosedMessage{c, err}) })
	h.connection.StartReceiving()
	return nil
}

func (h *connectionLogicHandler) elapsedTime() time.Duration {
	return time.Now().Sub(h.startTime)
}

func (h *connectionLogicHandler) tcpConnectionEstablished(msg message) error {
	m := msg.(*tcpConnectionEstablishedMessage)
	if h.state != connectionState_Connecting || h.connection != m.connection || m.connection.IsClosed() {
		log.Debugf("")
		return nil
	}

	h.heartbeatInfo = heartbeatInfo{h.packageNumber, true, h.elapsedTime()}

	if h.settings.DefaultUserCredentials != nil {
		h.connectingPhase = connectingPhase_Authentication
		h.authInfo = authInfo{uuid.NewV4(), h.elapsedTime()}
		h.connection.EnqueueSend(models.NewTcpPackage(
			models.Command_Authenticate,
			1,
			h.authInfo.CorrelationId,
			nil,
			h.settings.DefaultUserCredentials,
		))
	} else {
		h.goToConnectedState()
	}

	return nil
}

func (h *connectionLogicHandler) goToConnectedState() {
	h.state = connectionState_Connected
	h.connectingPhase = connectingPhase_Connected

	atomic.CompareAndSwapInt32(&h.wasConnected, 0, 1)

	h.connected.Raise(&models.ClientConnectionEventArgs{})

	if h.elapsedTime() - h.lastTimeoutsTimestamp >= h.settings.OperationTimeoutCheckPeriod() {
		h.operations.CheckTimeoutsAndRetry(h.connection)
		h.subscriptions.CheckTimeoutsAndRetry(h.connection)
		h.lastTimeoutsTimestamp = h.elapsedTime()
	}
}

func (h *connectionLogicHandler) tcpConnectionError(msg message) error {
	return nil
}

func (h *connectionLogicHandler) tcpConnectionClosed(msg message) error {
	return nil
}

func (h *connectionLogicHandler) handleTcpPackage(msg message) error {
	return nil
}

func (h *connectionLogicHandler) timerTick(msg message) error {
	switch h.state {
	case connectionState_Init:
		return nil
	case connectionState_Connecting:
		if h.connectingPhase == connectingPhase_Reconnecting && h.elapsedTime() - h.reconInfo.Timestamp >= h.settings.ReconnectionDelay() {
			log.Debug("TimerTick checking reconnection")

			h.reconInfo = reconnectionInfo{h.reconInfo.ReconnectionAttempt+1, h.elapsedTime()}
			if h.settings.MaxReconnections() >= 0 && h.reconInfo.ReconnectionAttempt > h.settings.MaxReconnections() {
				h.closeConnection(&closeConnectionMessage{"Reconnection limit reached.", nil})
			} else {
				h.raiseReconnecting()
				h.discoverEndpoint(nil)
			}
		}
		if h.connectingPhase == connectingPhase_Authentication && h.elapsedTime() - h.reconInfo.Timestamp >= h.settings.OperationTimeout() {
			h.raiseAuthFailed("Authentication timed out.")
			h.goToConnectedState()
		}
		if h.connectingPhase > connectingPhase_ConnectionEstablishing {
			h.manageHeartbeats()
		}
		return nil
	case connectionState_Connected:
		if h.elapsedTime() - h.lastTimeoutsTimestamp >= h.settings.OperationTimeoutCheckPeriod() {
			h.reconInfo = reconnectionInfo{0, h.elapsedTime()}
			h.operations.CheckTimeoutsAndRetry(h.connection)
			h.subscriptions.CheckTimeoutsAndRetry(h.connection)
			h.lastTimeoutsTimestamp = h.elapsedTime()
		}
		h.manageHeartbeats()
		return nil
	case connectionState_Closed:
		return nil
	default:
		return fmt.Errorf("Unknown state: %v", h.state)
	}
}

func (h *connectionLogicHandler) manageHeartbeats() {

}

func (h *connectionLogicHandler) Connected() models.EventHandlers { return h.connected }

func (h *connectionLogicHandler) Disconnected() models.EventHandlers { return h.disconnected }

func (h *connectionLogicHandler) Reconnecting() models.EventHandlers { return h.reconnecting }

func (h *connectionLogicHandler) raiseReconnecting() {
	h.reconnecting.Raise(&models.ClientReconnectingEventArgs{}) // TODO
}

func (h *connectionLogicHandler) Closed() models.EventHandlers { return h.closed }

func (h *connectionLogicHandler) raiseClosed(reason string) {
	h.closed.Raise(&models.ClientClosedEventArgs{}) // TODO
}

func (h *connectionLogicHandler) ErrorOccurred() models.EventHandlers { return h.errorOccurred }

func (h *connectionLogicHandler) raiseErrorOccurred(err error) {
	h.errorOccurred.Raise(&models.ClientErrorEventArgs{}) // TODO
}

func (h *connectionLogicHandler) AuthenticationFailed() models.EventHandlers { return h.authenticationFailed }

func (h *connectionLogicHandler) raiseAuthFailed(reason string) {
	h.authenticationFailed.Raise(&models.ClientAuthenticationFailedEventArgs{}) // TODO
}
