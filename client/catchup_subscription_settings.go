package client

import "fmt"

type CatchUpSubscriptionSettings struct {
	maxLiveQueueSize int
	readBatchSize    int
	verboseLogging   bool
	resolveLinkTos   bool
}

var CatchUpSubscriptionSettings_Default = &CatchUpSubscriptionSettings{CatchUpDefaultMaxPushQueueSize,
	CatchUpDefaultReadBatchSize, false, true}

func NewCatchUpSubscriptionSettings(
	maxLiveQueueSize int,
	readBatchSize int,
	verboseLogging bool,
	resolveLinkTos bool,
) *CatchUpSubscriptionSettings {
	if maxLiveQueueSize <= 0 {
		panic("maxLiveQueueSize should be positive")
	}
	if readBatchSize <= 0 {
		panic("readBatchSize should be positive")
	}
	if readBatchSize > MaxReadSize {
		panic(fmt.Sprintf("readBatchSize should be less than %d", MaxReadSize))
	}
	return &CatchUpSubscriptionSettings{
		maxLiveQueueSize: maxLiveQueueSize,
		readBatchSize:    readBatchSize,
		verboseLogging:   verboseLogging,
		resolveLinkTos:   resolveLinkTos,
	}
}

func (s *CatchUpSubscriptionSettings) MaxLiveQueueSize() int { return s.maxLiveQueueSize }

func (s *CatchUpSubscriptionSettings) ReadBatchSize() int { return s.readBatchSize }

func (s *CatchUpSubscriptionSettings) VerboseLogging() bool { return s.verboseLogging }

func (s *CatchUpSubscriptionSettings) ResolveLinkTos() bool { return s.resolveLinkTos }
