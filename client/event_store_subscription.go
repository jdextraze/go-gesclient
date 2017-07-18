package client

type EventStoreSubscription struct {
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
) *EventStoreSubscription {
	return &EventStoreSubscription{
		streamId:           streamId,
		lastCommitPosition: lastCommitPosition,
		lastEventNumber:    lastEventNumber,
		unsubscribe:        unsubscribe,
	}
}

func (s *EventStoreSubscription) IsSubscribedToAll() bool { return s.streamId == "" }

func (s *EventStoreSubscription) StreamId() string { return s.streamId }

func (s *EventStoreSubscription) LastCommitPosition() int64 { return s.lastCommitPosition }

func (s *EventStoreSubscription) LastEventNumber() *int { return s.lastEventNumber }

func (s *EventStoreSubscription) Close() error { return s.unsubscribe() }

func (s *EventStoreSubscription) Unsubscribe() error { return s.unsubscribe() }
