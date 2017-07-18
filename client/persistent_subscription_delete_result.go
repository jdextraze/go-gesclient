package client

type PersistentSubscriptionDeleteStatus int

const (
	PersistentSubscriptionDeleteStatus_Success = 0
	PersistentSubscriptionDeleteStatus_Failure = 1
)

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
