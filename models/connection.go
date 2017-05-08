package models

type Connection interface {
	Name() string

	//ConnectAsync() (<-chan struct{}, error)

	Close() error

	DeleteStreamAsync(stream string, expectedVersion int, hardDelete bool, userCredentials *UserCredentials) (
		<-chan *DeleteResult, error)

	AppendToStreamAsync(stream string, expectedVersion int, events []*EventData, userCredentials *UserCredentials) (
		<-chan *WriteResult, error)

	//StartTransactionAsync(stream string, expectedVersion int, userCredentials *UserCredentials) (
	//	<-chan Transaction, error)

	//ContinueTransaction(transactionId int64, userCredentials *UserCredentials) (Transaction, error)

	ReadEventAsync(stream string, eventNumber int, resolveTos bool, userCredentials *UserCredentials) (
		<-chan *EventReadResult, error)

	ReadStreamEventsForwardAsync(stream string, start int, max int, userCredentials *UserCredentials) (
		<-chan *StreamEventsSlice, error)

	ReadStreamEventsBackwardAsync(stream string, start int, max int, userCredentials *UserCredentials) (
		<-chan *StreamEventsSlice, error)

	ReadAllEventsForwardAsync(pos *Position, max int, resolveTos bool, userCredentials *UserCredentials) (
		<-chan *AllEventsSlice, error)

	ReadAllEventsBackwardAsync(pos *Position, max int, resolveTos bool, userCredentials *UserCredentials) (
		<-chan *AllEventsSlice, error)

	SubscribeToStreamAsync(
		stream string,
		resolveLinkTos bool,
		eventAppeared func(s Subscription, r ResolvedEvent),
		subscriptionDropped func(s Subscription, dr SubscriptionDropReason, err error),
		userCredentials *UserCredentials,
	) (<-chan Subscription, error)

	//SubscribeToStreamFrom(
	//	stream string,
	//	lastCheckpoint *int,
	//	catchupSubscriptionSettings CatchupSubscriptionSettings,
	//	eventAppeared func(s Subscription, r ResolvedEvent),
	//	liveProcessingStarted func(s Subscription),
	//	subscriptionDropped func(s Subscription, dr SubscriptionDropReason, err error),
	//	userCredentials *UserCredentials,
	//) (EventStoreStreamCatchUpSubscription, error)

	//SubscribeToAllAsync(
	//	resolveLinkTos bool,
	//	eventAppeared func(s Subscription, r ResolvedEvent),
	//	subscriptionDropped func(s Subscription, dr SubscriptionDropReason, err error),
	//	userCredentials *UserCredentials,
	//) (<-chan Subscription, error)

	ConnectToPersistentSubscriptionAsync(
		stream string,
		groupName string,
		eventAppeared func(s Subscription, r ResolvedEvent),
		subscriptionDropped func(s Subscription, dr SubscriptionDropReason, err error),
		userCredentials *UserCredentials,
		bufferSize int,
		autoAck bool,
	) (<-chan PersistentSubscription, error)

	//SubscribeToAllFrom(
	//	lastCheckpoint *Position,
	//	settings CatchupSubscriptionSettings,
	//	eventAppeared func(s Subscription, r ResolvedEvent),
	//	liveProcessingStarted func(s Subscription),
	//	subscriptionDropped func(s Subscription, dr SubscriptionDropReason, err error),
	//	userCredentials *UserCredentials,
	//) (EventStoreAllCatchUpSubscription, error)

	UpdatePersistentSubscriptionAsync(stream string, groupName string, settings PersistentSubscriptionSettings,
		userCredentials *UserCredentials) (<-chan *PersistentSubscriptionUpdateResult, error)

	CreatePersistentSubscriptionAsync(stream string, groupName string, settings PersistentSubscriptionSettings,
		userCredentials *UserCredentials) (<-chan *PersistentSubscriptionCreateResult, error)

	DeletePersistentSubscriptionAsync(stream string, groupName string,
		userCredentials *UserCredentials) (<-chan *PersistentSubscriptionDeleteResult, error)

	//SetStreamMetadataAsync(stream string, expectedMetastreamVersion int, metadata []byte,
	//	userCredentials *UserCredentials) (<-chan WriteResult, error)

	//GetStreamMetadataAsync(stream string, userCredentials *UserCredentials) (<-chan StreamMetadataResult, error)

	//SetSystemSettings(settings SystemSettings, userCredentials *UserCredentials) (<-chan struct{}, error)

	//Connected() EventHandlers
	//
	//Disconnected() EventHandlers
	//
	//Reconnecting() EventHandlers
	//
	//Closed() EventHandlers
	//
	//ErrorOccurred() EventHandlers
	//
	//AuthenticationFailed() EventHandlers

	Settings() *ConnectionSettings
}

type Event interface {}

type EventHandler func(evt Event) error

type EventHandlers interface {
	Add(EventHandler) error
	Remove(EventHandler) error
}
