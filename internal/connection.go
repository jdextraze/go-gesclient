package internal

import (
	"errors"
	"github.com/op/go-logging"
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/operations"
)

var log = logging.MustGetLogger("internal")

type connection struct {
	connectionSettings *models.ConnectionSettings
	clusterSettings    *models.ClusterSettings
	name               string
	endpointDiscoverer EndpointDiscoverer
	handler            ConnectionLogicHandler
}

func NewConnection(
	connectionSettings *models.ConnectionSettings,
	clusterSettings *models.ClusterSettings,
	endpointDiscoverer EndpointDiscoverer,
	name string,
) (models.Connection, error) {
	c := &connection{
		connectionSettings: connectionSettings,
		clusterSettings:    clusterSettings,
		endpointDiscoverer: endpointDiscoverer,
		name:               name,
	}
	handler, err := newConnectionLogicHandler(c, connectionSettings)
	c.handler = handler
	return c, err
}

func (c *connection) Name() string {
	return c.name
}

func (c *connection) Close() error {
	return c.handler.EnqueueMessage(newCloseConnectionMessage("Connection close requested by client.", nil))
}

func (c *connection) AppendToStreamAsync(
	stream string,
	expectedVersion int,
	events []*models.EventData,
	userCredentials *models.UserCredentials,
) (<-chan *models.WriteResult, error) {
	if stream == "" {
		panic("stream is empty")
	}
	if events == nil {
		panic("events is nil")
	}
	result := make(chan *models.WriteResult, 1)
	op := operations.NewAppendToStream(stream, events, expectedVersion, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) DeleteStreamAsync(
	stream string,
	expectedVersion int,
	hardDelete bool,
	userCredentials *models.UserCredentials,
) (<-chan *models.DeleteResult, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	result := make(chan *models.DeleteResult, 1)
	op := operations.NewDeleteStream(stream, expectedVersion, hardDelete, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) ReadEventAsync(
	stream string,
	eventNumber int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
) (<-chan *models.EventReadResult, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	result := make(chan *models.EventReadResult, 1)
	op := operations.NewReadEvent(stream, eventNumber, resolveTos, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) ReadStreamEventsForwardAsync(
	stream string,
	start int,
	max int,
	userCredentials *models.UserCredentials,
) (<-chan *models.StreamEventsSlice, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	result := make(chan *models.StreamEventsSlice, 1)
	op := operations.NewReadStreamEventsForward(stream, start, max, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) ReadStreamEventsBackwardAsync(
	stream string,
	start int,
	max int,
	userCredentials *models.UserCredentials,
) (<-chan *models.StreamEventsSlice, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	result := make(chan *models.StreamEventsSlice, 1)
	op := operations.NewReadStreamEventsBackward(stream, start, max, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) ReadAllEventsForwardAsync(
	position *models.Position,
	max int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
) (<-chan *models.AllEventsSlice, error) {
	if position == nil {
		panic("position is nil")
	}
	result := make(chan *models.AllEventsSlice, 1)
	op := operations.NewReadAllEventsForward(position, max, resolveTos, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) ReadAllEventsBackwardAsync(
	position *models.Position,
	max int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
) (<-chan *models.AllEventsSlice, error) {
	if position == nil {
		panic("position is nil")
	}
	result := make(chan *models.AllEventsSlice, 1)
	op := operations.NewReadAllEventsBackward(position, max, resolveTos, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) SubscribeToStreamAsync(
	stream string,
	resolveLinkTos bool,
	eventAppeared func(s models.Subscription, r models.ResolvedEvent),
	subscriptionDropped func(s models.Subscription, dr models.SubscriptionDropReason, err error),
	userCredentials *models.UserCredentials,
) (<-chan models.Subscription, error) {
	if stream == "" {
		panic("stream is empty")
	}
	if eventAppeared == nil {
		panic("eventAppeared is nil")
	}
	source := make(chan models.Subscription, 1)
	return source, c.handler.EnqueueMessage(&startSubscriptionMessage{
		source:              source,
		streamId:            stream,
		resolveLinkTos:      resolveLinkTos,
		userCredentials:     userCredentials,
		eventAppeared:       eventAppeared,
		subscriptionDropped: subscriptionDropped,
		maxRetries:          c.connectionSettings.MaxReconnections(),
		timeout:             c.connectionSettings.OperationTimeout(),
	})
}

func (c *connection) CreatePersistentSubscriptionAsync(
	stream string,
	groupName string,
	settings models.PersistentSubscriptionSettings,
	userCredentials *models.UserCredentials,
) (<-chan *models.PersistentSubscriptionCreateResult, error) {
	result := make(chan *models.PersistentSubscriptionCreateResult, 1)
	op := operations.NewCreatePersistentSubscription(stream, groupName, settings, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) UpdatePersistentSubscriptionAsync(
	stream string,
	groupName string,
	settings models.PersistentSubscriptionSettings,
	userCredentials *models.UserCredentials,
) (<-chan *models.PersistentSubscriptionUpdateResult, error) {
	result := make(chan *models.PersistentSubscriptionUpdateResult, 1)
	op := operations.NewUpdatePersistentSubscription(stream, groupName, settings, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) DeletePersistentSubscriptionAsync(
	stream string,
	groupName string,
	userCredentials *models.UserCredentials,
) (<-chan *models.PersistentSubscriptionDeleteResult, error) {
	result := make(chan *models.PersistentSubscriptionDeleteResult, 1)
	op := operations.NewDeletePersistentSubscription(stream, groupName, userCredentials, result)
	return result, c.EnqueueOperation(op)
}

func (c *connection) ConnectToPersistentSubscriptionAsync(
	stream string,
	groupName string,
	eventAppeared func(s models.Subscription, r models.ResolvedEvent),
	subscriptionDropped func(s models.Subscription, dr models.SubscriptionDropReason, err error),
	userCredentials *models.UserCredentials,
	bufferSize int,
	autoAck bool,
) (<-chan models.PersistentSubscription, error) {
	if groupName == "" {
		panic("groupName is empty")
	}
	if stream == "" {
		panic("stream is empty")
	}
	if eventAppeared == nil {
		panic("eventAppeared is nil")
	}
	result := make(chan models.PersistentSubscription, 1)
	op := operations.NewConnectToPersistentSubscription(stream, groupName, true, c, userCredentials, result)
	return op.Start()
}

func (c *connection) EnqueueOperation(op models.Operation) error {
	return c.handler.EnqueueMessage(&startOperationMessage{
		operation:  op,
		maxRetries: c.connectionSettings.MaxReconnections(),
		timeout:    c.connectionSettings.OperationTimeout(),
	})
}

func (c *connection) Settings() *models.ConnectionSettings {
	return c.connectionSettings
}
