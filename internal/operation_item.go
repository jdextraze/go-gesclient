package internal

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/satori/go.uuid"
	"sync/atomic"
	"time"
)

var operationItemNextSeqNo int64 = -1

type operationItem struct {
	seqNo         int64
	operation     client.Operation
	maxRetries    int
	timeout       time.Duration
	createdTime   time.Time
	ConnectionId  uuid.UUID
	CorrelationId uuid.UUID
	RetryCount    int
	LastUpdated   time.Time
}

func newOperationItem(operation client.Operation, maxRetries int, timeout time.Duration) *operationItem {
	return &operationItem{
		seqNo:         atomic.AddInt64(&operationItemNextSeqNo, 1),
		operation:     operation,
		maxRetries:    maxRetries,
		timeout:       timeout,
		createdTime:   time.Now().UTC(),
		CorrelationId: uuid.Must(uuid.NewV4()),
		RetryCount:    0,
		LastUpdated:   time.Now().UTC(),
	}
}

func (i *operationItem) String() string {
	return fmt.Sprintf("Operation %s (%s), retry count: %d, created: %s, last updated: %s.",
		i.operation, i.CorrelationId, i.RetryCount, i.createdTime, i.LastUpdated)
}
