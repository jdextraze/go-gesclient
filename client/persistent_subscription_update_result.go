package client

type PersistentSubscriptionUpdateStatus int

const (
	PersistentSubscriptionUpdateStatus_Success      = 0
	PersistentSubscriptionUpdateStatus_NotFound     = 1
	PersistentSubscriptionUpdateStatus_Failure      = 2
	PersistentSubscriptionUpdateStatus_AccessDenied = 3
)

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
