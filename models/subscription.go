package models

import "time"

type SubscriptionConfirmation struct {
	lastCommitPosition int64
	lastEventNumber    int32
}

func NewSubscriptionConfirmation(
	lastCommitPosition int64,
	lastEventNumber int32,
) *SubscriptionConfirmation {
	return &SubscriptionConfirmation{
		lastCommitPosition: lastCommitPosition,
		lastEventNumber:    lastEventNumber,
	}
}

func (c *SubscriptionConfirmation) LastCommitPosition() int64 {
	return c.lastCommitPosition
}

func (c *SubscriptionConfirmation) LastEventNumber() int32 {
	return c.lastEventNumber
}

type SubscriptionDropReason int

// TODO review this
const (
	SubscriptionDropReason_Error                         SubscriptionDropReason = -1
	SubscriptionDropReason_Unsubscribed                  SubscriptionDropReason = 0
	SubscriptionDropReason_AccessDenied                  SubscriptionDropReason = 1
	SubscriptionDropReason_NotFound                      SubscriptionDropReason = 2
	SubscriptionDropReason_PersistentSubscriptionDeleted SubscriptionDropReason = 3
	SubscriptionDropReason_SubscriberMaxCountReached     SubscriptionDropReason = 4

	SubscriptionDropReason_UserInitiated           SubscriptionDropReason = 5
	SubscriptionDropReason_CatchUpError            SubscriptionDropReason = 6
	SubscriptionDropReason_ProcessingQueueOverflow SubscriptionDropReason = 7
)

type SubscriptionDropped struct {
	Reason SubscriptionDropReason
	Error  error
}

type Subscription interface {
	Confirmation() *SubscriptionConfirmation
	Events() <-chan *ResolvedEvent
	Unsubscribe() error
	Dropped() <-chan *SubscriptionDropped
	Error() error
}

//

type PersistentSubscriptionNakEventAction int

const (
	PersistentSubscriptionNakEventAction_Unknown = 0
	PersistentSubscriptionNakEventAction_Park    = 1
	PersistentSubscriptionNakEventAction_Retry   = 2
	PersistentSubscriptionNakEventAction_Skip    = 3
	PersistentSubscriptionNakEventAction_Stop    = 4
)

type PersistentSubscription interface {
	Confirmation() *SubscriptionConfirmation
	Events() <-chan *ResolvedEvent
	Dropped() <-chan *SubscriptionDropped
	Acknowledge(events ...*ResolvedEvent)
	Failure(action PersistentSubscriptionNakEventAction, reason string, events ...*ResolvedEvent)
	Stop(timeout time.Duration) error
	Error() error
}
