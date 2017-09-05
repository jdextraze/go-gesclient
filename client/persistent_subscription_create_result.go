package client

type PersistentSubscriptionCreateStatus int

const (
	PersistentSubscriptionCreateStatus_Success  = 0
	PersistentSubscriptionCreateStatus_NotFound = 1
	PersistentSubscriptionCreateStatus_Failure  = 2
)

var persistentSubscriptionCreateStatuses = map[int]string{
	0: "Success",
	1: "NotFound",
	2: "Failure",
}

func (s PersistentSubscriptionCreateStatus) String() string {
	return persistentSubscriptionCreateStatuses[int(s)]
}

type PersistentSubscriptionCreateResult struct {
	status PersistentSubscriptionCreateStatus
}

func NewPersistentSubscriptionCreateResult(
	status PersistentSubscriptionCreateStatus,
) *PersistentSubscriptionCreateResult {
	return &PersistentSubscriptionCreateResult{
		status: status,
	}
}

func (r *PersistentSubscriptionCreateResult) GetStatus() PersistentSubscriptionCreateStatus {
	return r.status
}
