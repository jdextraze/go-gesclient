package gesclient

import "time"

type PersistentSubscriptionSettings struct {
	resolveLinkTos        bool
	startFrom             int32
	extraStatistics       bool
	messageTimeout        time.Duration
	MaxRetryCount         int32
	LiveBufferSize        int32
	ReadBatchSize         int32
	HistoryBufferSize     int32
	checkPointAfter       time.Duration
	minCheckPointCount    int32
	maxCheckPointCount    int32
	maxSubscriberCount    int32
	NamedConsumerStrategy SystemConsumerStrategies
}

func NewPersistentSubscriptionSettings(
	resolveLinkTos bool,
	startFrom int32,
	extraStatistics bool,
	messageTimeout time.Duration,
	maxRetryCount int32,
	liveBufferSize int32,
	readBatchSize int32,
	historyBufferSize int32,
	checkPointAfter time.Duration,
	minCheckPointCount int32,
	maxCheckPointCount int32,
	maxSubscriberCount int32,
	namedConsumerStrategy SystemConsumerStrategies,
) PersistentSubscriptionSettings {
	return PersistentSubscriptionSettings{
		resolveLinkTos:        resolveLinkTos,
		startFrom:             startFrom,
		extraStatistics:       extraStatistics,
		messageTimeout:        messageTimeout,
		MaxRetryCount:         maxRetryCount,
		LiveBufferSize:        liveBufferSize,
		ReadBatchSize:         readBatchSize,
		HistoryBufferSize:     historyBufferSize,
		checkPointAfter:       checkPointAfter,
		minCheckPointCount:    minCheckPointCount,
		maxCheckPointCount:    maxCheckPointCount,
		maxSubscriberCount:    maxSubscriberCount,
		NamedConsumerStrategy: namedConsumerStrategy,
	}
}

func (s *PersistentSubscriptionSettings) GetResolveLinkTos() bool { return s.resolveLinkTos }

func (s *PersistentSubscriptionSettings) GetStartFrom() int32 { return s.startFrom }

func (s *PersistentSubscriptionSettings) GetExtraStatistics() bool { return s.extraStatistics }

func (s *PersistentSubscriptionSettings) GetMessageTimeout() time.Duration { return s.messageTimeout }

func (s *PersistentSubscriptionSettings) GetCheckPointAfter() time.Duration { return s.checkPointAfter }

func (s *PersistentSubscriptionSettings) GetMinCheckPointCount() int32 { return s.minCheckPointCount }

func (s *PersistentSubscriptionSettings) GetMaxCheckPointCount() int32 { return s.maxCheckPointCount }

func (s *PersistentSubscriptionSettings) GetMaxSubscriberCount() int32 { return s.maxSubscriberCount }
