package internal

import (
	"github.com/jdextraze/go-gesclient/models"
	"net"
	"time"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"sync/atomic"
	"github.com/jdextraze/go-gesclient/subscriptions"
	"reflect"
	"github.com/jdextraze/go-gesclient/tasks"
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

var connectionState_values = []string{
	"Init",
	"Connecting",
	"Connected",
	"Closed",
}

func (s connectionState) String() string {
	return connectionState_values[s]
}

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
	timer                 *time.Ticker
	endpointDiscoverer    EndpointDiscoverer
	startTime             time.Time
	reconInfo             reconnectionInfo
	heartbeatInfo         heartbeatInfo
	authInfo              authInfo
	lastTimeoutsTimestamp time.Duration
	operations            *OperationsManager
	subscriptions         *SubscriptionsManager
	state                 connectionState
	connectingPhase       connectingPhase
	wasConnected          int32
	packageNumber         int
	connection            *models.PackageConnection
}

func NewConnectionLogicHandler(
	connection models.Connection,
	settings *models.ConnectionSettings,
) *connectionLogicHandler {
	if connection == nil {
		panic("connection is nil")
	}
	if settings == nil {
		panic("settings is nil")
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
		operations:           NewOperationsManager(connection.Name(), settings),
		subscriptions:        NewSubscriptionManager(connection.Name(), settings),
	}

	queue.RegisterHandler(&startConnectionMessage{}, obj.startConnection)
	queue.RegisterHandler(&closeConnectionMessage{}, obj.closeConnection)

	queue.RegisterHandler(&startOperationMessage{}, obj.startOperation)
	queue.RegisterHandler(&startSubscriptionMessage{}, obj.startSubscription)
	queue.RegisterHandler(&startPersistentSubscriptionMessage{}, obj.startPersistentSubscription)

	queue.RegisterHandler(&establishTcpConnectionMessage{}, obj.establishTcpConnection)
	queue.RegisterHandler(&tcpConnectionEstablishedMessage{}, obj.tcpConnectionEstablished)
	queue.RegisterHandler(&tcpConnectionErrorMessage{}, obj.tcpConnectionError)
	queue.RegisterHandler(&tcpConnectionClosedMessage{}, obj.tcpConnectionClosed)
	queue.RegisterHandler(&handleTcpPackageMessage{}, obj.handleTcpPackage)

	queue.RegisterHandler(&timerTickMessage{}, obj.timerTick)

	obj.timer = time.NewTicker(models.TimerPeriod)
	go func() {
		for range obj.timer.C {
			obj.EnqueueMessage(&timerTickMessage{})
		}
	}()

	return obj
}

func (h *connectionLogicHandler) TotalOperationCount() int {
	return h.operations.TotalOperationCount()
}

func (h *connectionLogicHandler) EnqueueMessage(msg message) error {
	_, isTimerTickMessage := msg.(*timerTickMessage)
	if h.settings.VerboseLogging() && !isTimerTickMessage {
		log.Debugf("enqueuing message %s", reflect.TypeOf(msg))
	}
	return h.queue.EnqueueMessage(msg)
}

func (h *connectionLogicHandler) startConnection(msg message) error {
	startConnectionMessage := msg.(*startConnectionMessage)
	if startConnectionMessage.task == nil {
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
		h.discoverEndpoint(startConnectionMessage.task)
	case connectionState_Connecting, connectionState_Connected:
		startConnectionMessage.task.SetError(fmt.Errorf(
			"EventStoreConnection '%s' is already active", h.esConnection.Name()))
	case connectionState_Closed:
		startConnectionMessage.task.SetError(fmt.Errorf(
			"EventStoreConnection '%s' is closed", h.esConnection.Name()))
	default:
		return fmt.Errorf("Unknown state '%v'", h.state)
	}
	return nil
}

func (h *connectionLogicHandler) discoverEndpoint(task *tasks.CompletionSource) {
	log.Debug("Discover endpoint")

	if h.state != connectionState_Connecting {
		return
	}
	if h.connectingPhase != connectingPhase_Reconnecting {
		return
	}

	h.connectingPhase = connectingPhase_EndpointDiscovery

	go func() {
		var remoteEndpoint net.Addr
		if h.connection != nil {
			remoteEndpoint = h.connection.RemoteEndpoint()
		}
		discovered := <-h.endpointDiscoverer.DiscoverAsync(remoteEndpoint)
		if discovered.error != nil {
			h.EnqueueMessage(newCloseConnectionMessage(
				"Failed to resolve TCP end point to which to connect.",
				discovered.error,
			))
			if task != nil {
				task.SetError(fmt.Errorf("Cannot resolve target endpoint"))
			}
			return
		}
		h.EnqueueMessage(newEstablishTcpConnectionMessage(discovered.nodeEndpoints))
		if task != nil {
			task.SetResult(nil)
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
		log.Debugf("StartOperation enqueue %s, %d, %s", m.operation, m.maxRetries, m.timeout)
		h.operations.EnqueueOperation(newOperationItem(m.operation, m.maxRetries, m.timeout))
	case connectionState_Connected:
		log.Debugf("StartOperation schedule %s, %d, %s", m.operation, m.maxRetries, m.timeout)
		h.operations.ScheduleOperation(newOperationItem(m.operation, m.maxRetries, m.timeout), h.connection)
	case connectionState_Closed:
		m.operation.Fail(fmt.Errorf("Connection %s is closed", h.esConnection.Name()))
	default:
		return fmt.Errorf("Unknown state: %s", h.state)
	}

	return nil
}

func (h *connectionLogicHandler) startSubscription(msg message) error {
	m := msg.(*startSubscriptionMessage)

	switch h.state {
	case connectionState_Init:
		m.source.SetError(fmt.Errorf("EventStoreConnection '%s' is not active.", h.esConnection.Name()))
	case connectionState_Connecting, connectionState_Connected:
		operation := subscriptions.NewVolatileSubscription(m.source, m.streamId, m.resolveLinkTos,
			m.userCredentials, m.eventAppeared, m.subscriptionDropped, h.settings.VerboseLogging(),
			func() (*models.PackageConnection, error) { return h.connection, nil })
		var state string
		if h.state == connectionState_Connected {
			state = "fire"
		} else {
			state = "enqueue"
		}
		log.Debugf("StartSubscription %s %s, %d, %s", state, operation, m.maxRetries, m.timeout)
		subscription := NewSubscriptionItem(operation, m.maxRetries, m.timeout)
		if h.state == connectionState_Connecting {
			h.subscriptions.EnqueueSubscription(subscription)
		} else {
			h.subscriptions.StartSubscription(subscription, h.connection)
		}
	case connectionState_Closed:
		m.source.SetError(fmt.Errorf("Object disposed: %s", h.esConnection.Name()))
	default:
		return fmt.Errorf("Unknown state: %s", h.state)
	}
	return nil
}

func (h *connectionLogicHandler) startPersistentSubscription(msg message) error {
	m := msg.(*startPersistentSubscriptionMessage)
	switch h.state {
	case connectionState_Init:
		m.source.SetError(fmt.Errorf("EventStoreConnection '%s' is not active.", h.esConnection.Name()))
	case connectionState_Connecting, connectionState_Connected:
		operation := subscriptions.NewConnectToPersistentSubscription(m.source, m.subscriptionId,
			m.bufferSize, m.streamId, m.userCredentials, m.eventAppeared, m.subscriptionDropped,
			h.settings.VerboseLogging(), func() (*models.PackageConnection, error) { return h.connection, nil })
		log.Debugf("StartSubscription %s %s, %d, %s", h.state, operation, m.maxRetries, m.timeout)
		subscription := NewSubscriptionItem(operation, m.maxRetries, m.timeout)
		if h.state == connectionState_Connecting {
			h.subscriptions.EnqueueSubscription(subscription)
		} else {
			h.subscriptions.StartSubscription(subscription, h.connection)
		}
	case connectionState_Closed:
		m.source.SetError(fmt.Errorf("Object disposed: %s", h.esConnection.Name()))
	default:
		return fmt.Errorf("Unknown state: %s", h.state)
	}
	return nil
}

func (h *connectionLogicHandler) establishTcpConnection(msg message) error {
	establishTcpConnection := msg.(*establishTcpConnectionMessage)
	var tcpEndpoint net.Addr
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
	h.connection = models.NewPackageConnection(log, tcpEndpoint, uuid.NewV4(), h.settings.UseSslConnection(),
		h.settings.TargetHost(), h.settings.ValidateService(), h.settings.ClientConnectionTimeout(),
		func(c *models.PackageConnection, p *models.Package) { h.EnqueueMessage(&handleTcpPackageMessage{c, p}) },
		func(c *models.PackageConnection, err error) { h.EnqueueMessage(&tcpConnectionErrorMessage{c, err}) },
		func(c *models.PackageConnection) { h.EnqueueMessage(&tcpConnectionEstablishedMessage{c}) },
		func(c *models.PackageConnection, err error) { h.EnqueueMessage(&tcpConnectionClosedMessage{c, err}) })
	return h.connection.StartReceiving()
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

	h.raiseConnected(h.connection.RemoteEndpoint())

	if h.elapsedTime()-h.lastTimeoutsTimestamp >= h.settings.OperationTimeoutCheckPeriod() {
		h.operations.CheckTimeoutsAndRetry(h.connection)
		h.subscriptions.CheckTimeoutsAndRetry(h.connection)
		h.lastTimeoutsTimestamp = h.elapsedTime()
	}
}

func (h *connectionLogicHandler) tcpConnectionError(msg message) error {
	m := msg.(*tcpConnectionErrorMessage)
	if h.connection != m.connection {
		return nil
	}
	if h.state == connectionState_Closed {
		return nil
	}
	log.Debugf("TcpConnectionError connId %s, exc %v", m.connection.ConnectionId(), m.error)
	return h.closeConnection(newCloseConnectionMessage("TCP connection error occurred.", m.error))
}

func (h *connectionLogicHandler) tcpConnectionClosed(msg message) error {
	m := msg.(*tcpConnectionClosedMessage)
	if h.state == connectionState_Init {
		return errors.New(":|")
	}
	if h.state == connectionState_Closed || h.connection != m.connection {
		var cid uuid.UUID
		if h.connection != nil {
			cid = h.connection.ConnectionId()
		}
		log.Debugf("IGNORED (_state: %s, _conn.ID: %s, conn.ID: %s: TCP connection to [%s, L%s] closed.",
			h.state, cid, m.connection.ConnectionId(), m.connection.RemoteEndpoint(), m.connection.LocalEndpoint())
		return nil
	}

	h.state = connectionState_Connecting
	h.connectingPhase = connectingPhase_Reconnecting

	log.Debugf("TCP connection to [%s, L%s, %s closed.", m.connection.RemoteEndpoint(),
		m.connection.LocalEndpoint(), m.connection.ConnectionId())

	h.subscriptions.PurgeSubscribedAndDroppedSubscriptions(h.connection.ConnectionId())
	h.reconInfo = reconnectionInfo{h.reconInfo.ReconnectionAttempt, h.elapsedTime()}

	if !atomic.CompareAndSwapInt32(&h.wasConnected, 1, 0) {
		h.raiseDisconnected(m.connection.RemoteEndpoint())
	}

	return nil
}

func (h *connectionLogicHandler) handleTcpPackage(msg message) error {
	m := msg.(*handleTcpPackageMessage)
	command := m.pkg.Command()
	correlationId := m.pkg.CorrelationId()

	if h.connection != m.connection || h.state == connectionState_Closed || h.state == connectionState_Init {
		log.Debugf("IGNORED: HandleTcpPackage connId %s, package %s, %s.", m.connection.ConnectionId(),
			command, correlationId)
		return nil
	}

	log.Debugf("HandleTcpPackage connId %s, package %s, %s.", m.connection.ConnectionId(), command,
		correlationId)
	h.packageNumber += 1

	if command == models.Command_HeartbeatResponseCommand {
		return nil
	}
	if command == models.Command_HeartbeatRequestCommand {
		return h.connection.EnqueueSend(models.NewTcpPackage(models.Command_HeartbeatResponseCommand, models.FlagsNone,
			correlationId, nil, nil))
	}

	if command == models.Command_Authenticated || command == models.Command_NotAuthenticated {
		if h.state == connectionState_Connecting &&
			h.connectingPhase == connectingPhase_Authentication &&
			h.authInfo.CorrelationId == correlationId {
			if command == models.Command_NotAuthenticated {
				h.raiseAuthFailed("Not authenticated")
			}
			h.goToConnectedState()
			return nil
		}
	}

	if command == models.Command_BadRequest && correlationId == uuid.Nil {
		message := string(m.pkg.Data())
		if message == "" {
			message = "<no message>"
		}
		return h.closeConnection(&closeConnectionMessage{
			reason: "Connection-wide bad request received. Too dangerous to continue.",
			error:  fmt.Errorf("Bad request received from server. Error: %s", message),
		})
	}

	if found, operation := h.operations.TryGetActiveOperation(correlationId); found {
		result := operation.operation.InspectPackage(m.pkg)
		log.Debugf("HandleTcpPackage OPERATION DECISION %s (%s), %s", result.Decision(), result.Description(),
			operation)
		switch result.Decision() {
		case models.InspectionDecision_DoNothing:
			break
		case models.InspectionDecision_EndOperation:
			h.operations.RemoveOperation(operation)
		case models.InspectionDecision_Retry:
			h.operations.ScheduleOperationRetry(operation)
		case models.InspectionDecision_Reconnect:
			h.reconnectTo(NewNodeEndpoints(result.TcpEndpoint(), result.SecureTcpEndpoint()))
			h.operations.ScheduleOperationRetry(operation)
		default:
			return fmt.Errorf("Unknown InspectionDecision: %s", result.Decision())
		}
		if h.state == connectionState_Connected {
			h.operations.TryScheduleWaitingOperations(m.connection)
		}
	} else if found, subscription := h.subscriptions.TryGetActiveSubscription(correlationId); found {
		result, err := subscription.Operation().InspectPackage(m.pkg)
		if err != nil {
			return err
		}
		log.Debugf("HandleTcpPackage SUBSCRIPTION DECISION %s (%s), %s", result.Decision(), result.Description(),
			subscription)
		switch result.Decision() {
		case models.InspectionDecision_DoNothing:
			break
		case models.InspectionDecision_EndOperation:
			h.subscriptions.RemoveSubscription(subscription)
		case models.InspectionDecision_Retry:
			h.subscriptions.ScheduleSubscriptionRetry(subscription)
		case models.InspectionDecision_Reconnect:
			h.reconnectTo(NewNodeEndpoints(result.TcpEndpoint(), result.SecureTcpEndpoint()))
			h.subscriptions.ScheduleSubscriptionRetry(subscription)
		case models.InspectionDecision_Subscribed:
			subscription.IsSubscribed = true
		default:
			return fmt.Errorf("Unknown InspectionDecision: %s", result.Decision())
		}
	}
	return nil
}

func (h *connectionLogicHandler) reconnectTo(endpoints *NodeEndpoints) {
	var endPoint net.Addr
	if h.settings.UseSslConnection() {
		if endpoints.SecureTcpEndpoint() == nil {
			endPoint = endpoints.TcpEndpoint()
		} else {
			endPoint = endpoints.SecureTcpEndpoint()
		}
	} else {
		endPoint = endpoints.TcpEndpoint()
	}
	if endPoint == nil {
		h.closeConnection("No end point is specified while trying to reconnect.")
		return
	}
	if h.state != connectionState_Connected || h.connection.RemoteEndpoint().String() == endPoint.String() {
		return
	}
	msg := fmt.Sprintf("EventStoreConnection '%s': going to reconnect to [%s]. Current endpoint: [%s, L%s].",
		h.esConnection.Name(), endPoint, h.connection.RemoteEndpoint(), h.connection.LocalEndpoint())
	if h.settings.VerboseLogging() {
		log.Debug(msg)
	}
	h.closeTcpConnection(msg)

	h.state = connectionState_Connecting
	h.connectingPhase = connectingPhase_EndpointDiscovery
	h.establishTcpConnection(endpoints)
}

func (h *connectionLogicHandler) timerTick(msg message) error {
	switch h.state {
	case connectionState_Init:
		return nil
	case connectionState_Connecting:
		if h.connectingPhase == connectingPhase_Reconnecting && h.elapsedTime()-h.reconInfo.Timestamp >= h.settings.ReconnectionDelay() {
			log.Debug("TimerTick checking reconnection")

			h.reconInfo = reconnectionInfo{h.reconInfo.ReconnectionAttempt + 1, h.elapsedTime()}
			if h.settings.MaxReconnections() >= 0 && h.reconInfo.ReconnectionAttempt > h.settings.MaxReconnections() {
				h.closeConnection(&closeConnectionMessage{"Reconnection limit reached.", nil})
			} else {
				h.raiseReconnecting()
				h.discoverEndpoint(nil)
			}
		}
		if h.connectingPhase == connectingPhase_Authentication && h.elapsedTime()-h.reconInfo.Timestamp >= h.settings.OperationTimeout() {
			h.raiseAuthFailed("Authentication timed out.")
			h.goToConnectedState()
		}
		if h.connectingPhase > connectingPhase_ConnectionEstablishing {
			return h.manageHeartbeats()
		}
		return nil
	case connectionState_Connected:
		if h.elapsedTime()-h.lastTimeoutsTimestamp >= h.settings.OperationTimeoutCheckPeriod() {
			h.reconInfo = reconnectionInfo{0, h.elapsedTime()}
			h.operations.CheckTimeoutsAndRetry(h.connection)
			h.subscriptions.CheckTimeoutsAndRetry(h.connection)
			h.lastTimeoutsTimestamp = h.elapsedTime()
		}
		return h.manageHeartbeats()
	case connectionState_Closed:
		return nil
	default:
		return fmt.Errorf("Unknown state: %v", h.state)
	}
}

func (h *connectionLogicHandler) manageHeartbeats() error {
	if h.connection == nil {
		return errors.New("")
	}

	var timeout time.Duration
	if h.heartbeatInfo.IsIntervalStage {
		timeout = h.settings.HeartbeatInterval()
	} else {
		timeout = h.settings.HeartbeatTimeout()
	}
	if h.elapsedTime()-h.heartbeatInfo.Timestamp < timeout {
		return nil
	}

	pkgNumber := h.packageNumber
	if h.heartbeatInfo.LastPackageNumber != pkgNumber {
		h.heartbeatInfo = heartbeatInfo{pkgNumber, true, h.elapsedTime()}
		return nil
	}

	if h.heartbeatInfo.IsIntervalStage {
		h.connection.EnqueueSend(models.NewTcpPackage(models.Command_HeartbeatRequestCommand, models.FlagsNone,
			uuid.NewV4(), nil, nil))
		h.heartbeatInfo = heartbeatInfo{h.heartbeatInfo.LastPackageNumber, false, h.elapsedTime()}
	} else {
		msg := fmt.Sprintf(
			"EventStoreConnection '%s': closing TCP connection [%s, %s, %s] due to HEARTBEAT TIMEOUT at pkgNum %d.",
			h.esConnection.Name(), h.connection.RemoteEndpoint(), h.connection.LocalEndpoint(),
			h.connection.ConnectionId(), pkgNumber)
		log.Info(msg)
		h.closeTcpConnection(msg)
	}
	return nil
}

func (h *connectionLogicHandler) Connected() models.EventHandlers { return h.connected }

func (h *connectionLogicHandler) raiseConnected(addr net.Addr) {
	h.connected.Raise(models.NewClientConnectionEventArgs(addr, h.esConnection))
}

func (h *connectionLogicHandler) Disconnected() models.EventHandlers { return h.disconnected }

func (h *connectionLogicHandler) raiseDisconnected(addr net.Addr) {
	h.disconnected.Raise(models.NewClientConnectionEventArgs(addr, h.esConnection))
}

func (h *connectionLogicHandler) Reconnecting() models.EventHandlers { return h.reconnecting }

func (h *connectionLogicHandler) raiseReconnecting() {
	h.reconnecting.Raise(models.NewClientReconnectingEventArgs(h.esConnection))
}

func (h *connectionLogicHandler) Closed() models.EventHandlers { return h.closed }

func (h *connectionLogicHandler) raiseClosed(reason string) {
	h.closed.Raise(models.NewClientClosedEventArgs(reason, h.esConnection))
}

func (h *connectionLogicHandler) ErrorOccurred() models.EventHandlers { return h.errorOccurred }

func (h *connectionLogicHandler) raiseErrorOccurred(err error) {
	h.errorOccurred.Raise(models.NewClientErrorEventArgs(err, h.esConnection))
}

func (h *connectionLogicHandler) AuthenticationFailed() models.EventHandlers { return h.authenticationFailed }

func (h *connectionLogicHandler) raiseAuthFailed(reason string) {
	h.authenticationFailed.Raise(models.NewClientAuthenticationFailedEventArgs(reason, h.esConnection))
}
