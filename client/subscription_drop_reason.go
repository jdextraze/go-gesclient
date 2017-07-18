package client

type SubscriptionDropReason int

const (
	SubscriptionDropReason_UserInitiated                 SubscriptionDropReason = 0
	SubscriptionDropReason_NotAuthenticated              SubscriptionDropReason = 1
	SubscriptionDropReason_AccessDenied                  SubscriptionDropReason = 2
	SubscriptionDropReason_SubscribingError              SubscriptionDropReason = 3
	SubscriptionDropReason_ServerError                   SubscriptionDropReason = 4
	SubscriptionDropReason_ConnectionClosed              SubscriptionDropReason = 5
	SubscriptionDropReason_CatchUpError                  SubscriptionDropReason = 6
	SubscriptionDropReason_ProcessingQueueOverflow       SubscriptionDropReason = 7
	SubscriptionDropReason_EventHandlerException         SubscriptionDropReason = 8
	SubscriptionDropReason_MaxSubscribersReached         SubscriptionDropReason = 9
	SubscriptionDropReason_PersistentSubscriptionDeleted SubscriptionDropReason = 10
	SubscriptionDropReason_NotFound                      SubscriptionDropReason = 11
	SubscriptionDropReason_Unknown                       SubscriptionDropReason = 100
)

var SubscriptionDropReason_name = map[int]string{
	0:   "UserInitiated",
	1:   "NotAuthenticated",
	2:   "AccessDenied",
	3:   "SubscriptionError",
	4:   "ServerError",
	5:   "ConnectionClosed",
	6:   "CatchUpError",
	7:   "ProcessingQueueOverflow",
	8:   "EventHandlerException",
	9:   "MaxSubscriberReached",
	10:  "PersistentSubscriptionDeleted",
	100: "NotFound",
}

func (r SubscriptionDropReason) String() string {
	return SubscriptionDropReason_name[int(r)]
}
