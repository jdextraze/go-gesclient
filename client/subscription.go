package client

type SubscriptionDropReason int

const (
	SubscriptionDropReason_UserInitiated    SubscriptionDropReason = 0
	SubscriptionDropReason_NotAuthenticated SubscriptionDropReason = 1
	SubscriptionDropReason_AccessDenied     SubscriptionDropReason = 2
	SubscriptionDropReason_SubscribingError SubscriptionDropReason = 3
	SubscriptionDropReason_ServerError      SubscriptionDropReason = 4
	SubscriptionDropReason_ConnectionClosed SubscriptionDropReason = 5

	SubscriptionDropReason_CatchUpError            SubscriptionDropReason = 6
	SubscriptionDropReason_ProcessingQueueOverflow SubscriptionDropReason = 7
	SubscriptionDropReason_EventHandlerException   SubscriptionDropReason = 8

	SubscriptionDropReason_MaxSubscribersReached SubscriptionDropReason = 9

	SubscriptionDropReason_PersistentSubscriptionDeleted SubscriptionDropReason = 10
	SubscriptionDropReason_Unknown                       SubscriptionDropReason = 100

	SubscriptionDropReason_NotFound SubscriptionDropReason = 11
)

var subscriptionDropReasonValues = []string{
	"Unsubscribed",
	"AccessDenied",
	"NotFound",
	"PersistentSubscriptionDeleted",
	"SubscriberMaxCountReached",
	"UserInitiated",
	"CatchUpError",
	"ProcessingQueueOverflow",
}

func (r SubscriptionDropReason) String() string {
	return subscriptionDropReasonValues[r]
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
