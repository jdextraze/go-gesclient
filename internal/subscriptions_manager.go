package internal

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/satori/go.uuid"
	"fmt"
	"time"
)

type SubscriptionsManager struct {
	connectionName            string
	settings                  *models.ConnectionSettings
	activeSubscriptions       map[uuid.UUID]*SubscriptionItem
	waitingSubscriptions      chan *SubscriptionItem
	retryPendingSubscriptions []*SubscriptionItem
}

func NewSubscriptionManager(connectionName string, settings *models.ConnectionSettings) *SubscriptionsManager {
	if settings == nil {
		panic("settings is nil")
	}
	return &SubscriptionsManager{
		connectionName:            connectionName,
		settings:                  settings,
		activeSubscriptions:       map[uuid.UUID]*SubscriptionItem{},
		waitingSubscriptions:      make(chan *SubscriptionItem, 4096), // TODO buffer size
		retryPendingSubscriptions: []*SubscriptionItem{},
	}
}

func (m *SubscriptionsManager) TryGetActiveSubscription(correlationId uuid.UUID) (bool, *SubscriptionItem) {
	item, ok := m.activeSubscriptions[correlationId]
	return ok, item
}

func (m *SubscriptionsManager) CleanUp() {
	err := fmt.Errorf("Connection '%s' was closed", m.connectionName)
	for i, s := range m.activeSubscriptions {
		s.Operation().DropSubscription(models.SubscriptionDropReason_ConnectionClosed, err, nil)
		delete(m.activeSubscriptions, i)
	}
	for len(m.waitingSubscriptions) > 0 {
		s := <-m.waitingSubscriptions
		s.Operation().DropSubscription(models.SubscriptionDropReason_ConnectionClosed, err, nil)
	}
	for i, s := range m.retryPendingSubscriptions {
		s.Operation().DropSubscription(models.SubscriptionDropReason_ConnectionClosed, err, nil)
		m.retryPendingSubscriptions[i] = nil
	}
	m.retryPendingSubscriptions = []*SubscriptionItem{}
}

func (m *SubscriptionsManager) PurgeSubscribedAndDroppedSubscriptions(connectionId uuid.UUID) {
	for _, s := range m.activeSubscriptions {
		if s.IsSubscribed && uuid.Equal(s.ConnectionId, connectionId) {
			s.Operation().ConnectionClosed()
		}
	}
	for i, s := range m.activeSubscriptions {
		if s.IsSubscribed && uuid.Equal(s.ConnectionId, connectionId) {
			delete(m.activeSubscriptions, i)
		}
	}
}

func (m *SubscriptionsManager) CheckTimeoutsAndRetry(c *models.PackageConnection) {
	if c == nil {
		panic("connection is nil")
	}

	removeSubscriptions := []*SubscriptionItem{}
	retrySubscriptions := []*SubscriptionItem{}
	for _, s := range m.activeSubscriptions {
		if s.IsSubscribed {
			continue
		}
		if s.ConnectionId != c.ConnectionId() {
			m.retryPendingSubscriptions = append(m.retryPendingSubscriptions, s)
		} else if s.Timeout() > time.Duration(0) && time.Now().UTC().Sub(s.LastUpdated) > m.settings.OperationTimeout() {
			err := fmt.Errorf("EventStoreConnection '%s': subscription never got confirmation from server.\n"+
				"UTC now: %s, operation: %s.", m.connectionName, time.Now().UTC(), s)
			log.Errorf("%v", err)

			if m.settings.FailOnNoServerResponse() {
				s.Operation().DropSubscription(models.SubscriptionDropReason_SubscribingError, err, nil)
				removeSubscriptions = append(removeSubscriptions, s)
			} else {
				retrySubscriptions = append(retrySubscriptions, s)
			}
		}
	}

	for _, s := range retrySubscriptions {
		m.ScheduleSubscriptionRetry(s)
	}
	for _, s := range removeSubscriptions {
		m.RemoveSubscription(s)
	}

	if len(m.retryPendingSubscriptions) > 0 {
		for _, s := range m.retryPendingSubscriptions {
			s.RetryCount += 1
			m.StartSubscription(s, c)
		}
		m.retryPendingSubscriptions = []*SubscriptionItem{}
	}

	for len(m.waitingSubscriptions) > 0 {
		m.StartSubscription(<-m.waitingSubscriptions, c)
	}
}

func (m *SubscriptionsManager) EnqueueSubscription(s *SubscriptionItem) {
	m.waitingSubscriptions <- s
}

func (m *SubscriptionsManager) StartSubscription(s *SubscriptionItem, c *models.PackageConnection) {
	if c == nil {
		panic("connection is nil")
	}

	if s.IsSubscribed {
		m.logDebug("StartSubscription REMOVING due to already subscribed %s.", s)
		m.RemoveSubscription(s)
		return
	}

	s.CorrelationId = uuid.NewV4()
	s.ConnectionId = c.ConnectionId()
	s.LastUpdated = time.Now().UTC()

	m.activeSubscriptions[s.CorrelationId] = s

	ok, err := s.Operation().Subscribe(s.CorrelationId, c)
	if err != nil {
		// TODO handle error
		return
	} else if !ok {
		m.logDebug("StartSubscription REMOVING AS COULD NOT SUBSCRIBE %s.", s)
	} else {
		m.logDebug("StartSubscription SUBSCRIBING %s.", s)
	}
}

func (m *SubscriptionsManager) RemoveSubscription(s *SubscriptionItem) bool {
	_, found := m.activeSubscriptions[s.CorrelationId]
	m.logDebug("RemoveSubscription %s, result: %v", s, found)
	if !found {
		return false
	}
	delete(m.activeSubscriptions, s.CorrelationId)
	return true
}

func (m *SubscriptionsManager) ScheduleSubscriptionRetry(s *SubscriptionItem) {
	if !m.RemoveSubscription(s) {
		m.logDebug("RemoveSubscription failed when trying to retry %s", s)
		return
	}

	if s.MaxRetries() >= 0 && s.RetryCount >= s.MaxRetries() {
		m.logDebug("RETRIES LIMIT REACHED when trying to retry %s", s)
		s.Operation().DropSubscription(models.SubscriptionDropReason_SubscribingError,
			fmt.Errorf("Retries limit of %d reached for %s", s.MaxRetries(), s), nil)
		return
	}

	m.logDebug("Retrying subscription %s.", s)
	m.retryPendingSubscriptions = append(m.retryPendingSubscriptions, s)
}

func (m *SubscriptionsManager) logDebug(format string, args ...interface{}) {
	if m.settings.VerboseLogging() {
		log.Debugf(format, args...)
	}
}
