package client

type PersistentSubscriptionCreateStatus int

const (
	PersistentSubscriptionCreateStatus_Success  = 0
	PersistentSubscriptionCreateStatus_NotFound = 1
	PersistentSubscriptionCreateStatus_Failure  = 2
)

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
