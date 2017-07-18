package client

import "github.com/satori/go.uuid"

type ConnectToPersistentSubscriptions interface {
	NotifyEventsProcessed(processedEvents []uuid.UUID) error
	NotifyEventsFailed(processedEvents []uuid.UUID, action PersistentSubscriptionNakEventAction, reason string) error
	Unsubscribe() error
}

type PersistentEventStoreSubscription struct {
	*EventStoreSubscription
	subscriptionOperation ConnectToPersistentSubscriptions
}

func NewPersistentEventStoreSubscription(
	subscriptionOperation ConnectToPersistentSubscriptions,
	streamId string,
	lastCommitPosition int64,
	lastEventNumber *int,
) *PersistentEventStoreSubscription {
	obj := &PersistentEventStoreSubscription{
		subscriptionOperation: subscriptionOperation,
	}
	obj.EventStoreSubscription = NewEventStoreSubscription(streamId, lastCommitPosition, lastEventNumber,
		obj.unsubscribe)
	return obj
}

func (s *PersistentEventStoreSubscription) unsubscribe() error {
	return s.subscriptionOperation.Unsubscribe()
}

func (s *PersistentEventStoreSubscription) NotifyEventsProcessed(processedEvents []uuid.UUID) error {
	return s.subscriptionOperation.NotifyEventsProcessed(processedEvents)
}

func (s *PersistentEventStoreSubscription) NotifyEventsFailed(
	processedEvents []uuid.UUID,
	action PersistentSubscriptionNakEventAction,
	reason string,
) error {
	return s.subscriptionOperation.NotifyEventsFailed(processedEvents, action, reason)
}
