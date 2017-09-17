package client

type PersistentSubscriptionUpdateStatus int

const (
	PersistentSubscriptionUpdateStatus_Success      = 0
	PersistentSubscriptionUpdateStatus_NotFound     = 1
	PersistentSubscriptionUpdateStatus_Failure      = 2
	PersistentSubscriptionUpdateStatus_AccessDenied = 3
)

var persistentSubscriptionUpdateStatuses = map[int]string{
	0: "Success",
	1: "NotFound",
	2: "Failure",
	3: "AccessDenied",
}

func (s PersistentSubscriptionUpdateStatus) String() string {
	return persistentSubscriptionCreateStatuses[int(s)]
}

type PersistentSubscriptionUpdateResult struct {
	status PersistentSubscriptionUpdateStatus
}

func NewPersistentSubscriptionUpdateResult(
	status PersistentSubscriptionUpdateStatus,
) *PersistentSubscriptionUpdateResult {
	return &PersistentSubscriptionUpdateResult{
		status: status,
	}
}

func (r *PersistentSubscriptionUpdateResult) GetStatus() PersistentSubscriptionUpdateStatus {
	return r.status
}
