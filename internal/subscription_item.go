package internal

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/subscriptions"
	"github.com/satori/go.uuid"
	"time"
)

type SubscriptionItem struct {
	operation     subscriptions.Subscription
	maxRetries    int
	timeout       time.Duration
	createdTime   time.Time
	ConnectionId  uuid.UUID
	CorrelationId uuid.UUID
	IsSubscribed  bool
	RetryCount    int
	LastUpdated   time.Time
}

func NewSubscriptionItem(
	operation subscriptions.Subscription,
	maxRetries int,
	timeout time.Duration,
) *SubscriptionItem {
	if operation == nil {
		panic("operation is nil")
	}
	return &SubscriptionItem{
		operation:     operation,
		maxRetries:    maxRetries,
		timeout:       timeout,
		createdTime:   time.Now().UTC(),
		CorrelationId: uuid.Must(uuid.NewV4()),
		RetryCount:    0,
		LastUpdated:   time.Now().UTC(),
	}
}

func (i *SubscriptionItem) Operation() subscriptions.Subscription { return i.operation }

func (i *SubscriptionItem) MaxRetries() int { return i.maxRetries }

func (i *SubscriptionItem) Timeout() time.Duration { return i.timeout }

func (i *SubscriptionItem) CreatedTime() time.Time { return i.createdTime }

func (i *SubscriptionItem) String() string {
	return fmt.Sprintf("Subscription %s (%s): %s, is subscribed: %v, retry count: %d, created: %s, last updated: %s",
		i.operation, i.CorrelationId, i.operation, i.IsSubscribed, i.RetryCount, i.createdTime, i.LastUpdated)
}
