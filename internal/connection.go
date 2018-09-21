package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/common"
	"github.com/jdextraze/go-gesclient/operations"
	"github.com/jdextraze/go-gesclient/subscriptions"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
	"time"
)

type connection struct {
	connectionSettings *client.ConnectionSettings
	clusterSettings    *client.ClusterSettings
	name               string
	endpointDiscoverer EndpointDiscoverer
	handler            ConnectionLogicHandler
}

func NewConnection(
	settings *client.ConnectionSettings,
	clusterSettings *client.ClusterSettings,
	endpointDiscoverer EndpointDiscoverer,
	name string,
) client.Connection {
	if settings == nil {
		panic("settings is nil")
	}
	if endpointDiscoverer == nil {
		panic("endpointDiscoverer is nil")
	}

	if name == "" {
		name = fmt.Sprintf("ES-%s", uuid.Must(uuid.NewV4()))
	}

	c := &connection{
		connectionSettings: settings,
		clusterSettings:    clusterSettings,
		endpointDiscoverer: endpointDiscoverer,
		name:               name,
	}
	c.handler = NewConnectionLogicHandler(c, settings)
	return c
}

func (c *connection) Name() string {
	return c.name
}

func (c *connection) ConnectAsync() *tasks.Task {
	source := tasks.NewCompletionSource()
	c.handler.EnqueueMessage(newStartConnectionMessage(source, c.endpointDiscoverer))
	return source.Task()
}

func (c *connection) Close() error {
	return c.handler.EnqueueMessage(newCloseConnectionMessage("Connection close requested by client.", nil))
}

func (c *connection) DeleteStreamAsync(
	stream string,
	expectedVersion int,
	hardDelete bool,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewDeleteStream(source, stream, expectedVersion, hardDelete, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) AppendToStreamAsync(
	stream string,
	expectedVersion int,
	events []*client.EventData,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		panic("stream is empty")
	}
	if events == nil {
		panic("events is nil")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewAppendToStream(source, c.connectionSettings.RequireMaster(), stream, expectedVersion, events,
		userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) StartTransactionAsync(
	stream string,
	expectedVersion int,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		panic("stream is empty")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewStartTransaction(source, c.connectionSettings.RequireMaster(), stream, expectedVersion, c,
		userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) ContinueTransaction(
	transactionId int64,
	userCredentials *client.UserCredentials,
) *client.Transaction {
	return client.NewTransaction(transactionId, userCredentials, c)
}

func (c *connection) TransactionalWriteAsync(
	transaction *client.Transaction,
	events []*client.EventData,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if transaction == nil {
		panic("transaction is nil")
	}
	if events == nil {
		panic("events is nil")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewTransactionalWrite(source, c.connectionSettings.RequireMaster(), transaction.TransactionId(),
		events, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) CommitTransactionAsync(
	transaction *client.Transaction,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if transaction == nil {
		panic("transaction is nil")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewCommitTransaction(source, c.connectionSettings.RequireMaster(), transaction.TransactionId(),
		userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) ReadEventAsync(
	stream string,
	eventNumber int,
	resolveTos bool,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewReadEvent(source, stream, eventNumber, resolveTos, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) ReadStreamEventsForwardAsync(
	stream string,
	start int,
	max int,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewReadStreamEventsForward(source, stream, start, max, resolveLinkTos,
		c.Settings().RequireMaster(), userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) ReadStreamEventsBackwardAsync(
	stream string,
	start int,
	max int,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		return nil, errors.New("stream must be present")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewReadStreamEventsBackward(source, stream, start, max, resolveLinkTos,
		c.Settings().RequireMaster(), userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) ReadAllEventsForwardAsync(
	position *client.Position,
	max int,
	resolveTos bool,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if position == nil {
		panic("position is nil")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewReadAllEventsForward(source, position, max, resolveTos, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) ReadAllEventsBackwardAsync(
	position *client.Position,
	max int,
	resolveTos bool,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if position == nil {
		panic("position is nil")
	}
	source := tasks.NewCompletionSource()
	op := operations.NewReadAllEventsBackward(source, position, max, resolveTos, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) SubscribeToStreamAsync(
	stream string,
	resolveLinkTos bool,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		panic("stream is empty")
	}
	if eventAppeared == nil {
		panic("eventAppeared is nil")
	}
	source := tasks.NewCompletionSource()
	return source.Task(), c.handler.EnqueueMessage(newStartSubscriptionMessage(source, stream, resolveLinkTos,
		userCredentials, eventAppeared, subscriptionDropped, c.connectionSettings.MaxReconnections(),
		c.connectionSettings.OperationTimeout()))
}

func (c *connection) SubscribeToAllAsync(
	resolveLinkTos bool,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if eventAppeared == nil {
		panic("eventAppeared is nil")
	}
	source := tasks.NewCompletionSource()
	return source.Task(), c.handler.EnqueueMessage(newStartSubscriptionMessage(source, "", resolveLinkTos,
		userCredentials, eventAppeared, subscriptionDropped, c.Settings().MaxRetries(),
		c.Settings().OperationTimeout()))
}

func (c *connection) SubscribeToStreamFrom(
	stream string,
	lastCheckpoint *int,
	settings *client.CatchUpSubscriptionSettings,
	eventAppeared client.CatchUpEventAppearedHandler,
	liveProcessingStarted client.LiveProcessingStartedHandler,
	subscriptionDropped client.CatchUpSubscriptionDroppedHandler,
	userCredentials *client.UserCredentials,
) (client.CatchUpSubscription, error) {
	sub := subscriptions.NewStreamCatchUpSubscription(c, stream, lastCheckpoint, userCredentials, eventAppeared,
		liveProcessingStarted, subscriptionDropped, settings)
	sub.Start()
	return sub, nil
}

func (c *connection) SubscribeToAllFrom(
	lastCheckpoint *client.Position,
	settings *client.CatchUpSubscriptionSettings,
	eventAppeared client.CatchUpEventAppearedHandler,
	liveProcessingStarted client.LiveProcessingStartedHandler,
	subscriptionDropped client.CatchUpSubscriptionDroppedHandler,
	userCredentials *client.UserCredentials,
) (client.CatchUpSubscription, error) {
	sub := subscriptions.NewAllCatchUpSubscription(c, lastCheckpoint, userCredentials, eventAppeared,
		liveProcessingStarted, subscriptionDropped, settings)
	sub.Start()
	return sub, nil
}

func (c *connection) ConnectToPersistentSubscriptionAsync(
	stream string,
	groupName string,
	eventAppeared client.PersistentEventAppearedHandler,
	subscriptionDropped client.PersistentSubscriptionDroppedHandler,
	userCredentials *client.UserCredentials,
	bufferSize int,
	autoAck bool,
) (*tasks.Task, error) {
	sub := NewPersistentSubscription(groupName, stream, eventAppeared, subscriptionDropped,
		userCredentials, c.Settings(), c.handler, bufferSize, autoAck)
	return sub.Start(), nil
}

func (c *connection) CreatePersistentSubscriptionAsync(
	stream string,
	groupName string,
	settings *client.PersistentSubscriptionSettings,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	source := tasks.NewCompletionSource()
	op := operations.NewCreatePersistentSubscription(source, stream, groupName, settings, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) UpdatePersistentSubscriptionAsync(
	stream string,
	groupName string,
	settings *client.PersistentSubscriptionSettings,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	source := tasks.NewCompletionSource()
	op := operations.NewUpdatePersistentSubscription(source, stream, groupName, settings, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) DeletePersistentSubscriptionAsync(
	stream string,
	groupName string,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	source := tasks.NewCompletionSource()
	op := operations.NewDeletePersistentSubscription(source, stream, groupName, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) SetStreamMetadataAsync(
	stream string,
	expectedMetastreamVersion int,
	metadata interface{},
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	if stream == "" {
		panic("stream is empty")
	}
	if common.SystemStreams_IsMetastream(stream) {
		panic(fmt.Errorf("Setting metadata for metastream '%s' is not supported.", stream))
	}
	source := tasks.NewCompletionSource()
	var metaevent *client.EventData
	switch metadata.(type) {
	case []byte:
		metaevent = client.NewEventData(uuid.Must(uuid.NewV4()), common.SystemEventTypes_StreamMetadata, true, metadata.([]byte),
			nil)
	case client.StreamMetadata, *client.StreamMetadata:
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metaevent = client.NewEventData(uuid.Must(uuid.NewV4()), common.SystemEventTypes_StreamMetadata, true, data, nil)
	default:
		return nil, fmt.Errorf("Unknown metadata type: %v", metadata)
	}
	op := operations.NewAppendToStream(source, c.Settings().RequireMaster(), common.SystemStreams_MetastreamOf(stream),
		expectedMetastreamVersion, []*client.EventData{metaevent}, userCredentials)
	return source.Task(), c.enqueueOperation(op)
}

func (c *connection) GetStreamMetadataAsync(
	stream string,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	t, err := c.ReadEventAsync(common.SystemStreams_MetastreamOf(stream), -1, false, userCredentials)
	if err != nil {
		return nil, err
	}
	return t.ContinueWith(func(t *tasks.Task) (interface{}, error) {
		if t.Error() != nil {
			return nil, t.Error()
		}
		res := t.Result().(*client.EventReadResult)
		switch res.Status() {
		case client.EventReadStatus_Success:
			if res.Event() == nil {
				return nil, errors.New("Event is nil while operation result is Success.")
			}
			evt := res.Event().OriginalEvent()
			if evt == nil || evt.Data() == nil || len(evt.Data()) == 0 {
				return client.NewStreamMetadataResult(res.Stream(), false, -1, client.StreamMetadata{}), nil
			}
			if metadata, err := client.StreamMetadataFromJsonBytes(evt.Data()); err != nil {
				return nil, err
			} else {
				return client.NewStreamMetadataResult(res.Stream(), false, -1, metadata), nil
			}
		case client.EventReadStatus_NotFound, client.EventReadStatus_NoStream:
			return client.NewStreamMetadataResult(res.Stream(), false, -1, client.StreamMetadata{}), nil
		case client.EventReadStatus_StreamDeleted:
			return client.NewStreamMetadataResult(res.Stream(), true, 2147483647, client.StreamMetadata{}), nil
		default:
			return nil, fmt.Errorf("Unexpected ReadEventResult: %v", res.Status())
		}
	}), nil
}

func (c *connection) SetSystemSettings(
	settings *client.SystemSettings,
	userCredentials *client.UserCredentials,
) (*tasks.Task, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	evt := client.NewEventData(uuid.Must(uuid.NewV4()), common.SystemEventTypes_Settings, true, data, nil)
	return c.AppendToStreamAsync(common.SystemStreams_SettingsStream, client.ExpectedVersion_Any,
		[]*client.EventData{evt}, userCredentials)
}

func (c *connection) enqueueOperation(op client.Operation) error {
	for {
		if c.handler.TotalOperationCount() <= c.Settings().MaxQueueSize() {
			break
		}
		time.Sleep(time.Millisecond)
	}
	return c.handler.EnqueueMessage(newStartOperationMessage(op, c.connectionSettings.MaxReconnections(),
		c.connectionSettings.OperationTimeout()))
}

func (c *connection) Settings() *client.ConnectionSettings {
	return c.connectionSettings
}

func (c *connection) Connected() client.EventHandlers { return c.handler.Connected() }

func (c *connection) Disconnected() client.EventHandlers { return c.handler.Disconnected() }

func (c *connection) Reconnecting() client.EventHandlers { return c.handler.Reconnecting() }

func (c *connection) Closed() client.EventHandlers { return c.handler.Closed() }

func (c *connection) ErrorOccurred() client.EventHandlers { return c.handler.ErrorOccurred() }

func (c *connection) AuthenticationFailed() client.EventHandlers {
	return c.handler.AuthenticationFailed()
}

func (c *connection) String() string {
	return fmt.Sprintf(
		"Connection{name: '%s' connectionSettings: %+v clusterSettings: %+v}",
		c.name, c.connectionSettings, c.clusterSettings,
	)
}
