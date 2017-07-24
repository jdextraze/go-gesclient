package client

import (
	"time"
)

type PersistentSubscription interface {
	Acknowledge(events []ResolvedEvent) error
	Fail(events []ResolvedEvent, action PersistentSubscriptionNakEventAction, reason string) error
	Stop(timeout ...time.Duration) error
}
