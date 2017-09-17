package client

type PersistentSubscriptionDeleteStatus int

const (
	PersistentSubscriptionDeleteStatus_Success = 0
	PersistentSubscriptionDeleteStatus_Failure = 1
)

var persistentSubscriptionDeleteStatuses = map[int]string{
	0: "Success",
	1: "Failure",
}

func (s PersistentSubscriptionDeleteStatus) String() string {
	return persistentSubscriptionCreateStatuses[int(s)]
}

type PersistentSubscriptionDeleteResult struct {
	status PersistentSubscriptionDeleteStatus
}

func NewPersistentSubscriptionDeleteResult(
	status PersistentSubscriptionDeleteStatus,
) *PersistentSubscriptionDeleteResult {
	return &PersistentSubscriptionDeleteResult{
		status: status,
	}
}

func (r *PersistentSubscriptionDeleteResult) GetStatus() PersistentSubscriptionDeleteStatus {
	return r.status
}
