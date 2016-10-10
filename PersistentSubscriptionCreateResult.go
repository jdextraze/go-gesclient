package gesclient

type PersistentSubscriptionCreateStatus int

const (
	PersistentSubscriptionCreateStatus_Success  = 0
	PersistentSubscriptionCreateStatus_NotFound = 1
	PersistentSubscriptionCreateStatus_Failure  = 2
)

type PersistentSubscriptionCreateResult struct {
	status PersistentSubscriptionCreateStatus
	error  error
}

func newPersistentSubscriptionCreateResult(
	status PersistentSubscriptionCreateStatus,
	error error,
) *PersistentSubscriptionCreateResult {
	return &PersistentSubscriptionCreateResult{
		status: status,
		error:  error,
	}
}

func (r *PersistentSubscriptionCreateResult) GetStatus() PersistentSubscriptionCreateStatus {
	return r.status
}

func (r *PersistentSubscriptionCreateResult) GetError() error { return r.error }
