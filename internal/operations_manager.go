package internal

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/satori/go.uuid"
	"sort"
	"sync"
	"time"
)

type OperationsManager struct {
	connectionName         string
	settings               *client.ConnectionSettings
	activeOperations       map[uuid.UUID]*operationItem
	waitingOperations      chan *operationItem
	retryPendingOperations []*operationItem
	lock                   sync.Locker
	totalOperationCount    int
}

func NewOperationsManager(
	connectionName string,
	settings *client.ConnectionSettings,
) *OperationsManager {
	if settings == nil {
		panic("settings is nil")
	}
	return &OperationsManager{
		connectionName:         connectionName,
		settings:               settings,
		activeOperations:       map[uuid.UUID]*operationItem{},
		waitingOperations:      make(chan *operationItem, 4096), // TODO buffer size
		retryPendingOperations: []*operationItem{},
		lock:                &sync.Mutex{},
		totalOperationCount: 0,
	}
}

func (m *OperationsManager) TryGetActiveOperation(correlationId uuid.UUID) (bool, *operationItem) {
	item, ok := m.activeOperations[correlationId]
	return ok, item
}

func (m *OperationsManager) CleanUp() {
	err := fmt.Errorf("Connection '%s' was closed", m.connectionName)
	for i, o := range m.activeOperations {
		o.operation.Fail(err)
		delete(m.activeOperations, i)
	}
	for len(m.waitingOperations) > 0 {
		o := <-m.waitingOperations
		o.operation.Fail(err)
	}
	for i, o := range m.retryPendingOperations {
		o.operation.Fail(err)
		m.retryPendingOperations[i] = nil
	}
	m.retryPendingOperations = []*operationItem{}
	m.totalOperationCount = 0
}

func (m *OperationsManager) CheckTimeoutsAndRetry(c *client.PackageConnection) {
	if c == nil {
		panic("connection is nil")
	}

	removeOperations := []*operationItem{}
	retryOperations := []*operationItem{}
	for _, o := range m.activeOperations {
		if o.ConnectionId != c.ConnectionId() {
			retryOperations = append(retryOperations, o)
		} else if o.timeout > time.Duration(0) && time.Now().UTC().Sub(o.LastUpdated) > m.settings.OperationTimeout() {
			err := fmt.Errorf("EventStoreConnection '%s': operation never got response from server.\n"+
				"UTC now: %s, operation: %s.", m.connectionName, time.Now().UTC(), o)
			log.Errorf("%v", err)

			if m.settings.FailOnNoServerResponse() {
				o.operation.Fail(err)
				removeOperations = append(removeOperations, o)
			} else {
				retryOperations = append(retryOperations, o)
			}
		}
	}

	for _, s := range retryOperations {
		m.ScheduleOperationRetry(s)
	}
	for _, s := range removeOperations {
		m.RemoveOperation(s)
	}

	if len(m.retryPendingOperations) > 0 {
		sort.Sort(BySeqNo(m.retryPendingOperations))
		for _, s := range m.retryPendingOperations {
			oldCorrId := s.CorrelationId
			s.CorrelationId = uuid.NewV4()
			s.RetryCount += 1
			log.Debugf("retrying, old corrId: %s, operation %s.", oldCorrId, s)
			m.ScheduleOperation(s, c)
		}
		m.retryPendingOperations = []*operationItem{}
	}

	m.TryScheduleWaitingOperations(c)
}

func (m *OperationsManager) TryScheduleWaitingOperations(c *client.PackageConnection) {
	if c == nil {
		panic("connection is nil")
	}
	m.lock.Lock()
	for len(m.waitingOperations) > 0 && len(m.activeOperations) < m.settings.MaxConcurrentItem() {
		m.ExecuteOperation(<-m.waitingOperations, c) // TODO handler error
	}
	m.totalOperationCount = len(m.activeOperations) + len(m.waitingOperations)
	m.lock.Unlock()
}

func (m *OperationsManager) ExecuteOperation(o *operationItem, c *client.PackageConnection) error {
	o.ConnectionId = c.ConnectionId()
	o.LastUpdated = time.Now().UTC()
	m.activeOperations[o.CorrelationId] = o

	pkg, err := o.operation.CreateNetworkPackage(o.CorrelationId)
	if err != nil {
		return err
	}
	m.logDebug("ExecuteOperation package %s, %s, %s.", pkg.Command(), pkg.CorrelationId(), o)
	return c.EnqueueSend(pkg)
}

func (m *OperationsManager) TotalOperationCount() int { return m.totalOperationCount }

func (m *OperationsManager) ScheduleOperationRetry(o *operationItem) {
	if !m.RemoveOperation(o) {
		m.logDebug("RemoveSubscription failed when trying to retry %s", o)
		return
	}

	if o.maxRetries >= 0 && o.RetryCount >= o.maxRetries {
		m.logDebug("RETRIES LIMIT REACHED when trying to retry %s", o)
		o.operation.Fail(fmt.Errorf("Retries limit of %d reached for %s", o.maxRetries, o))
		return
	}

	m.logDebug("Retrying subscription %s.", o)
	m.retryPendingOperations = append(m.retryPendingOperations, o)
}

func (m *OperationsManager) RemoveOperation(o *operationItem) bool {
	_, found := m.activeOperations[o.CorrelationId]
	m.logDebug("RemoveSubscription %s, result: %v", o, found)
	if !found {
		return false
	}
	delete(m.activeOperations, o.CorrelationId)
	return true
}

func (m *OperationsManager) EnqueueOperation(operation *operationItem) error {
	m.logDebug("EnqueueOperation WAITING for %s", operation)
	m.waitingOperations <- operation
	return nil
}

func (m *OperationsManager) ScheduleOperation(operation *operationItem, conn *client.PackageConnection) error {
	m.waitingOperations <- operation
	m.TryScheduleWaitingOperations(conn)
	return nil
}

func (m *OperationsManager) logDebug(format string, args ...interface{}) {
	if m.settings.VerboseLogging() {
		log.Debugf(format, args...)
	}
}

type BySeqNo []*operationItem

func (a BySeqNo) Len() int           { return len(a) }
func (a BySeqNo) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeqNo) Less(i, j int) bool { return a[i].seqNo < a[j].seqNo }
