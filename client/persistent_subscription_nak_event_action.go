package client

type PersistentSubscriptionNakEventAction int

const (
	PersistentSubscriptionNakEventAction_Unknown PersistentSubscriptionNakEventAction = 0
	PersistentSubscriptionNakEventAction_Park    PersistentSubscriptionNakEventAction = 1
	PersistentSubscriptionNakEventAction_Retry   PersistentSubscriptionNakEventAction = 2
	PersistentSubscriptionNakEventAction_Skip    PersistentSubscriptionNakEventAction = 3
	PersistentSubscriptionNakEventAction_Stop    PersistentSubscriptionNakEventAction = 4
)

var PersistentSubscriptionNakEventAction_name = map[int]string{
	0: "Unknown",
	1: "Park",
	2: "Retry",
	3: "Skip",
	4: "Stop",
}

func (x PersistentSubscriptionNakEventAction) String() string {
	return PersistentSubscriptionNakEventAction_name[int(x)]
}
