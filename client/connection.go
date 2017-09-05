package client

import (
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type EventAppearedHandler func(s EventStoreSubscription, r *ResolvedEvent) error

type SubscriptionDroppedHandler func(s EventStoreSubscription, dr SubscriptionDropReason, err error) error

type CatchUpEventAppearedHandler func(s CatchUpSubscription, r *ResolvedEvent) error

type CatchUpSubscriptionDroppedHandler func(s CatchUpSubscription, dr SubscriptionDropReason, err error) error

type LiveProcessingStartedHandler func(s CatchUpSubscription) error

type PersistentEventAppearedHandler func(s PersistentSubscription, r *ResolvedEvent) error

type PersistentSubscriptionDroppedHandler func(s PersistentSubscription, dr SubscriptionDropReason, err error) error

type Connection interface {
	Name() string

	// Use Task.Wait()
	ConnectAsync() *tasks.Task

	Close() error

	// Task.Result() returns *client.DeleteResult
	DeleteStreamAsync(stream string, expectedVersion int, hardDelete bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	// Task.Result() returns *client.WriteResult
	AppendToStreamAsync(stream string, expectedVersion int, events []*EventData, userCredentials *UserCredentials) (
		*tasks.Task, error)

	//StartTransactionAsync(stream string, expectedVersion int, userCredentials *UserCredentials) (
	//	<-chan Transaction, error)

	//ContinueTransaction(transactionId int64, userCredentials *UserCredentials) (Transaction, error)

	// Task.Result() returns *client.EventReadResult
	ReadEventAsync(stream string, eventNumber int, resolveTos bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	// Task.Result() returns *client.StreamEventsSlice
	ReadStreamEventsForwardAsync(stream string, start int, max int, resolveLinkTos bool,
		userCredentials *UserCredentials) (*tasks.Task, error)

	// Task.Result() returns *client.StreamEventsSlice
	ReadStreamEventsBackwardAsync(stream string, start int, max int, resolveLinkTos bool,
		userCredentials *UserCredentials) (*tasks.Task, error)

	// Task.Result() returns *client.AllEventsSlice
	ReadAllEventsForwardAsync(pos *Position, max int, resolveTos bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	// Task.Result() returns *client.AllEventsSlice
	ReadAllEventsBackwardAsync(pos *Position, max int, resolveTos bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	// Task.Result() returns client.EventStoreSubscription
	SubscribeToStreamAsync(
		stream string,
		resolveLinkTos bool,
		eventAppeared EventAppearedHandler,
		subscriptionDropped SubscriptionDroppedHandler,
		userCredentials *UserCredentials,
	) (*tasks.Task, error)

	SubscribeToStreamFrom(
		stream string,
		lastCheckpoint *int,
		catchupSubscriptionSettings *CatchUpSubscriptionSettings,
		eventAppeared CatchUpEventAppearedHandler,
		liveProcessingStarted LiveProcessingStartedHandler,
		subscriptionDropped CatchUpSubscriptionDroppedHandler,
		userCredentials *UserCredentials,
	) (CatchUpSubscription, error)

	// Task.Result() returns client.EventStoreSubscription
	SubscribeToAllAsync(
		resolveLinkTos bool,
		eventAppeared EventAppearedHandler,
		subscriptionDropped SubscriptionDroppedHandler,
		userCredentials *UserCredentials,
	) (*tasks.Task, error)

	// Task.Result() returns client.PersistentSubscription
	ConnectToPersistentSubscriptionAsync(
		stream string,
		groupName string,
		eventAppeared PersistentEventAppearedHandler,
		subscriptionDropped PersistentSubscriptionDroppedHandler,
		userCredentials *UserCredentials,
		bufferSize int,
		autoAck bool,
	) (*tasks.Task, error)

	SubscribeToAllFrom(
		lastCheckpoint *Position,
		settings *CatchUpSubscriptionSettings,
		eventAppeared CatchUpEventAppearedHandler,
		liveProcessingStarted LiveProcessingStartedHandler,
		subscriptionDropped CatchUpSubscriptionDroppedHandler,
		userCredentials *UserCredentials,
	) (CatchUpSubscription, error)

	// Task.Result() returns *client.PersistentSubscriptionUpdateResult
	UpdatePersistentSubscriptionAsync(stream string, groupName string, settings *PersistentSubscriptionSettings,
		userCredentials *UserCredentials) (*tasks.Task, error)

	// Task.Result() returns *client.PersistentSubscriptionCreateResult
	CreatePersistentSubscriptionAsync(stream string, groupName string, settings *PersistentSubscriptionSettings,
		userCredentials *UserCredentials) (*tasks.Task, error)

	// Task.Result() returns *client.PersistentSubscriptionDeleteResult
	DeletePersistentSubscriptionAsync(stream string, groupName string,
		userCredentials *UserCredentials) (*tasks.Task, error)

	// Task.Result() returns *client.WriteResult
	SetStreamMetadataAsync(stream string, expectedMetastreamVersion int, metadata interface{},
		userCredentials *UserCredentials) (*tasks.Task, error)

	// Task.Result() returns *client.StreamMetadataResult
	GetStreamMetadataAsync(stream string, userCredentials *UserCredentials) (*tasks.Task, error)

	//SetSystemSettings(settings SystemSettings, userCredentials *UserCredentials) (<-chan struct{}, error)

	Connected() EventHandlers

	Disconnected() EventHandlers

	Reconnecting() EventHandlers

	Closed() EventHandlers

	ErrorOccurred() EventHandlers

	AuthenticationFailed() EventHandlers

	Settings() *ConnectionSettings
}

type CatchUpSubscription interface {
	Stop(timeout ...time.Duration) error
}
