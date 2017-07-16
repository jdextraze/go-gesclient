package client

import "time"

const (
	DefaultMaxQueueSize        int = 5000
	DefaultMaxConcurrentItems  int = 5000
	DefaultMaxOperationRetries int = 10
	DefaultMaxReconnections    int = 10

	DefaultRequireMaster bool = true

	DefaultReconnectionDelay           time.Duration = 100 * time.Millisecond
	DefaultOperationTimeout            time.Duration = 7 * time.Second
	DefaultOperationTimeoutCheckPeriod time.Duration = time.Second

	TimerPeriod                           time.Duration = 200 * time.Millisecond
	MaxReadSize                           int           = 4096
	DefaultMaxClusterDiscoverAttempts     int           = 10
	DefaultClusterManagerExternalHttpPort int           = 30778

	CatchUpDefaultReadBatchSize    int = 500
	CatchUpDefaultMaxPushQueueSize int = 10000
)
