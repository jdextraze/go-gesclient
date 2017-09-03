package client

import (
	"fmt"
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
) *PersistentSubscriptionSettings {
	return &PersistentSubscriptionSettings{
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

func (s *PersistentSubscriptionSettings) String() string {
	return fmt.Sprintf(
		"PersistentSubscriptionSettings{resolveLinkTos: %t startFrom: %d extraStatistics: %t mesageTimeout: %s "+
			"maxRetryCount: %d liveBufferSize: %d readBatchSize: %d historyBufferSize: %d checkPointAfter: %s"+
			"minCheckPointCount: %d maxCheckPointCount: %d maxSubscriberCount: %d namedConsumerStrategy: %s}",
		s.resolveLinkTos, s.startFrom, s.extraStatistics, s.messageTimeout, s.MaxRetryCount, s.LiveBufferSize,
		s.ReadBatchSize, s.HistoryBufferSize, s.checkPointAfter, s.minCheckPointCount, s.maxCheckPointCount,
		s.maxSubscriberCount, s.NamedConsumerStrategy,
	)
}

var DefaultPersistentSubscriptionSettings = NewPersistentSubscriptionSettings(false, -1, false, 30*time.Second,
	500, 500, 10, 20, 2*time.Second, 10, 1000, 0, common.SystemConsumerStrategies_RoundRobin)
