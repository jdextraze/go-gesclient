package client

type EventStoreSubscription interface {
	IsSubscribedToAll() bool
	StreamId() string
	LastCommitPosition() int64
	LastEventNumber() *int
	Close() error
	Unsubscribe() error
}

type eventStoreSubscription struct {
	streamId           string
	lastCommitPosition int64
	lastEventNumber    *int
	unsubscribe        func() error
}

func NewEventStoreSubscription(
	streamId string,
	lastCommitPosition int64,
	lastEventNumber *int,
	unsubscribe func() error,
) *eventStoreSubscription {
	return &eventStoreSubscription{
		streamId:           streamId,
		lastCommitPosition: lastCommitPosition,
		lastEventNumber:    lastEventNumber,
		unsubscribe:        unsubscribe,
	}
}

func (s *eventStoreSubscription) IsSubscribedToAll() bool { return s.streamId == "" }

func (s *eventStoreSubscription) StreamId() string { return s.streamId }

func (s *eventStoreSubscription) LastCommitPosition() int64 { return s.lastCommitPosition }

func (s *eventStoreSubscription) LastEventNumber() *int { return s.lastEventNumber }

func (s *eventStoreSubscription) Close() error { return s.unsubscribe() }

func (s *eventStoreSubscription) Unsubscribe() error { return s.unsubscribe() }
