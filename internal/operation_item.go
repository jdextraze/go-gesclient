package internal

import (
	"github.com/jdextraze/go-gesclient/models"
	"time"
	"sync/atomic"
	"fmt"
	"github.com/satori/go.uuid"
)

var operationItemNextSeqNo int64 = -1

type operationItem struct {
	seqNo int64
	operation models.Operation
	maxRetries int
	timeout time.Duration
	createdTime time.Time
	ConnectionId uuid.UUID
	CorrelationId uuid.UUID
	RetryCount int
	LastUpdated time.Time
}

func newOperationItem(operation models.Operation, maxRetries int, timeout time.Duration) *operationItem {
	return &operationItem{
		seqNo: atomic.AddInt64(&operationItemNextSeqNo, 1),
		operation: operation,
		maxRetries: maxRetries,
		timeout: timeout,
		createdTime: time.Now().UTC(),
		CorrelationId: uuid.NewV4(),
		RetryCount: 0,
		LastUpdated: time.Now().UTC(),
	}
}

func (i *operationItem) String() string {
	return fmt.Sprintf("Operation %s (%s), retry count: %d, created: %s, last updated: %s.",
		i.operation, i.CorrelationId, i.RetryCount, i.createdTime, i.LastUpdated)
}
