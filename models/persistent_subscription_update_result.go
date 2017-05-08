package models

type PersistentSubscriptionUpdateStatus int

const (
	PersistentSubscriptionUpdateStatus_Success      = 0
	PersistentSubscriptionUpdateStatus_NotFound     = 1
	PersistentSubscriptionUpdateStatus_Failure      = 2
	PersistentSubscriptionUpdateStatus_AccessDenied = 3
)

type PersistentSubscriptionUpdateResult struct {
	status PersistentSubscriptionUpdateStatus
	error  error
}

func NewPersistentSubscriptionUpdateResult(
	status PersistentSubscriptionUpdateStatus,
	error error,
) *PersistentSubscriptionUpdateResult {
	return &PersistentSubscriptionUpdateResult{
		status: status,
		error:  error,
	}
}

func (r *PersistentSubscriptionUpdateResult) GetStatus() PersistentSubscriptionUpdateStatus {
	return r.status
}

func (r *PersistentSubscriptionUpdateResult) GetError() error { return r.error }
