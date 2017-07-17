package client

type PersistentSubscriptionNakEventAction int

const (
	PersistentSubscriptionNakEventAction_Unknown PersistentSubscriptionNakEventAction = iota
	PersistentSubscriptionNakEventAction_Park
	PersistentSubscriptionNakEventAction_Retry
	PersistentSubscriptionNakEventAction_Skip
	PersistentSubscriptionNakEventAction_Stop
)

var PersistentSubscriptionNakEventAction_names = []string{
	"Unknown",
	"Park",
	"Retry",
	"Skip",
	"Stop",
}

func (x PersistentSubscriptionNakEventAction) String() string {
	return PersistentSubscriptionNakEventAction_names[x]
}
