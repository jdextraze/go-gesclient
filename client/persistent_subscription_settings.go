package client

import (
	"github.com/jdextraze/go-gesclient/common"
	"time"
)

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
	NamedConsumerStrategy common.SystemConsumerStrategies
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
	namedConsumerStrategy common.SystemConsumerStrategies,
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

func (s *PersistentSubscriptionSettings) ResolveLinkTos() bool { return s.resolveLinkTos }

func (s *PersistentSubscriptionSettings) StartFrom() int32 { return s.startFrom }

func (s *PersistentSubscriptionSettings) ExtraStatistics() bool { return s.extraStatistics }

func (s *PersistentSubscriptionSettings) MessageTimeout() time.Duration { return s.messageTimeout }

func (s *PersistentSubscriptionSettings) CheckPointAfter() time.Duration { return s.checkPointAfter }

func (s *PersistentSubscriptionSettings) MinCheckPointCount() int32 { return s.minCheckPointCount }

func (s *PersistentSubscriptionSettings) MaxCheckPointCount() int32 { return s.maxCheckPointCount }

func (s *PersistentSubscriptionSettings) MaxSubscriberCount() int32 { return s.maxSubscriberCount }
