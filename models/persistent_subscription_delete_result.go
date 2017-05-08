package models

type PersistentSubscriptionDeleteStatus int

const (
	PersistentSubscriptionDeleteStatus_Success = 0
	PersistentSubscriptionDeleteStatus_Failure = 1
)

type PersistentSubscriptionDeleteResult struct {
	status PersistentSubscriptionDeleteStatus
	error  error
}

func NewPersistentSubscriptionDeleteResult(
	status PersistentSubscriptionDeleteStatus,
	error error,
) *PersistentSubscriptionDeleteResult {
	return &PersistentSubscriptionDeleteResult{
		status: status,
		error:  error,
	}
}

func (r *PersistentSubscriptionDeleteResult) GetStatus() PersistentSubscriptionDeleteStatus {
	return r.status
}

func (r *PersistentSubscriptionDeleteResult) GetError() error { return r.error }
