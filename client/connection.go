package client

import (
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type EventAppearedHandler func(s *EventStoreSubscription, r *ResolvedEvent) error

type SubscriptionDroppedHandler func(s *EventStoreSubscription, dr SubscriptionDropReason, err error) error

type CatchUpEventAppearedHandler func(s CatchUpSubscription, r *ResolvedEvent) error

type CatchUpSubscriptionDroppedHandler func(s CatchUpSubscription, dr SubscriptionDropReason, err error) error

type LiveProcessingStartedHandler func(s CatchUpSubscription) error

type Connection interface {
	Name() string

	ConnectAsync() *tasks.Task

	Close() error

	DeleteStreamAsync(stream string, expectedVersion int, hardDelete bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	AppendToStreamAsync(stream string, expectedVersion int, events []*EventData, userCredentials *UserCredentials) (
		*tasks.Task, error)

	//StartTransactionAsync(stream string, expectedVersion int, userCredentials *UserCredentials) (
	//	<-chan Transaction, error)

	//ContinueTransaction(transactionId int64, userCredentials *UserCredentials) (Transaction, error)

	ReadEventAsync(stream string, eventNumber int, resolveTos bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	ReadStreamEventsForwardAsync(stream string, start int, max int, resolveLinkTos bool,
		userCredentials *UserCredentials) (*tasks.Task, error)

	ReadStreamEventsBackwardAsync(stream string, start int, max int, resolveLinkTos bool,
		userCredentials *UserCredentials) (*tasks.Task, error)

	ReadAllEventsForwardAsync(pos *Position, max int, resolveTos bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

	ReadAllEventsBackwardAsync(pos *Position, max int, resolveTos bool, userCredentials *UserCredentials) (
		*tasks.Task, error)

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

	SubscribeToAllAsync(
		resolveLinkTos bool,
		eventAppeared EventAppearedHandler,
		subscriptionDropped SubscriptionDroppedHandler,
		userCredentials *UserCredentials,
	) (*tasks.Task, error)

	//ConnectToPersistentSubscriptionAsync(
	//	stream string,
	//	groupName string,
	//	eventAppeared EventAppearedHandler,
	//	subscriptionDropped SubscriptionDroppedHandler,
	//	userCredentials *UserCredentials,
	//	bufferSize int,
	//	autoAck bool,
	//) (<-chan PersistentSubscription, error)

	SubscribeToAllFrom(
		lastCheckpoint *Position,
		settings *CatchUpSubscriptionSettings,
		eventAppeared CatchUpEventAppearedHandler,
		liveProcessingStarted LiveProcessingStartedHandler,
		subscriptionDropped CatchUpSubscriptionDroppedHandler,
		userCredentials *UserCredentials,
	) (CatchUpSubscription, error)

	UpdatePersistentSubscriptionAsync(stream string, groupName string, settings *PersistentSubscriptionSettings,
		userCredentials *UserCredentials) (*tasks.Task, error)

	CreatePersistentSubscriptionAsync(stream string, groupName string, settings *PersistentSubscriptionSettings,
		userCredentials *UserCredentials) (*tasks.Task, error)

	DeletePersistentSubscriptionAsync(stream string, groupName string,
		userCredentials *UserCredentials) (*tasks.Task, error)

	//SetStreamMetadataAsync(stream string, expectedMetastreamVersion int, metadata []byte,
	//	userCredentials *UserCredentials) (<-chan WriteResult, error)

	//GetStreamMetadataAsync(stream string, userCredentials *UserCredentials) (<-chan StreamMetadataResult, error)

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
