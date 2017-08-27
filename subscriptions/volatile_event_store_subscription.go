package subscriptions

import (
	"github.com/jdextraze/go-gesclient/client"
)

type VolatileEventStoreSubscription struct {
	client.EventStoreSubscription
	subscriptionOperation *VolatileSubscription
}

func NewVolatileEventStoreSubscription(
	subscriptionOperation *VolatileSubscription,
	streamId string,
	lastCommitPosition int64,
	lastEventNumber *int,
) *VolatileEventStoreSubscription {
	obj := &VolatileEventStoreSubscription{
		subscriptionOperation: subscriptionOperation,
	}
	obj.EventStoreSubscription = client.NewEventStoreSubscription(streamId, lastCommitPosition, lastEventNumber,
		obj.unsubscribe)
	return obj
}

func (s *VolatileEventStoreSubscription) unsubscribe() error {
	return s.subscriptionOperation.Unsubscribe()
}
