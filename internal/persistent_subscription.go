package internal

import (
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"sync"
	"time"
	"errors"
	"github.com/satori/go.uuid"
	log "github.com/jdextraze/go-gesclient/logger"
	"unsafe"
	"sync/atomic"
)

var dropSubscriptionEvent = client.NewResolvedEvent(nil)

type dropData struct {
	reason client.SubscriptionDropReason
	err    error
}

type persistentSubscription struct {
	subscriptionId      string
	streamId            string
	eventAppeared       client.PersistentEventAppearedHandler
	subscriptionDropped client.PersistentSubscriptionDroppedHandler
	userCredentials     *client.UserCredentials
	settings            *client.ConnectionSettings
	handler             ConnectionLogicHandler
	bufferSize          int
	autoAck             bool
	subscription        *client.PersistentEventStoreSubscription
	queue               chan *client.ResolvedEvent
	isProcessing        int32
	dropData            *dropData
	isDropped           int32
	stopped             sync.WaitGroup
}

func NewPersistentSubscription(
	subscriptionId string,
	streamId string,
	eventAppeared client.PersistentEventAppearedHandler,
	subscriptionDropped client.PersistentSubscriptionDroppedHandler,
	userCredentials *client.UserCredentials,
	settings *client.ConnectionSettings,
	handler ConnectionLogicHandler,
	bufferSize int,
	autoAck bool,
) *persistentSubscription {
	obj := &persistentSubscription{
		subscriptionId:      subscriptionId,
		streamId:            streamId,
		eventAppeared:       eventAppeared,
		subscriptionDropped: subscriptionDropped,
		userCredentials:     userCredentials,
		settings:            settings,
		handler:             handler,
		bufferSize:          bufferSize,
		autoAck:             autoAck,
	}
	return obj
}

func (s *persistentSubscription) Start() *tasks.Task {
	s.stopped.Done()
	source := tasks.NewCompletionSource()
	s.handler.EnqueueMessage(NewStartPersistentSubscriptionMessage(source, s.subscriptionId, s.streamId,
		s.bufferSize, s.userCredentials, s.onEventAppeared, s.onSubscriptionDropped, s.settings.MaxRetries(),
		s.settings.OperationTimeout()))
	return source.Task().ContinueWith(func(t *tasks.Task) error {
		s.subscription = &client.PersistentEventStoreSubscription{}
		return t.Result(s.subscription)
	})
}

func (s *persistentSubscription) Acknowledge(events []client.ResolvedEvent) error {
	if len(events) > 2000 {
		return errors.New("events is limited to 2000 to ack at a time")
	}
	ids := make([]uuid.UUID, len(events))
	for i, e := range events {
		ids[i] = e.OriginalEvent().EventId()
	}
	return s.subscription.NotifyEventsProcessed(ids)
}

func (s *persistentSubscription) Fail(
	events []client.ResolvedEvent,
	action client.PersistentSubscriptionNakEventAction,
	reason string,
) error {
	if len(events) > 2000 {
		return errors.New("events is limited to 2000 to ack at a time")
	}
	ids := make([]uuid.UUID, len(events))
	for i, e := range events {
		ids[i] = e.OriginalEvent().EventId()
	}
	return s.subscription.NotifyEventsFailed(ids, action, reason)
}

func (s *persistentSubscription) Stop(timeout ...time.Duration) (err error) {
	if len(timeout) > 1 {
		panic("invalid number of arguments")
	}
	if s.settings.VerboseLogging() {
		log.Debugf("Persistent Subscription to %s: requesting stop...", s.streamId)
	}
	s.enqueueSubscriptionDropNotification(client.SubscriptionDropReason_UserInitiated, nil)
	if len(timeout) == 0 {
		return
	}
	go func() {
		<-time.After(timeout[0])
		err = errors.New("Could not stop in time")
		s.stopped.Done()
	}()
	s.stopped.Wait()
	return
}

func (s *persistentSubscription) enqueueSubscriptionDropNotification(reason client.SubscriptionDropReason, err error) {
	dropData := dropData{reason, err}
	ptr := unsafe.Pointer(s.dropData)
	if atomic.CompareAndSwapPointer(&ptr, nil, unsafe.Pointer(&dropData)) {
		s.enqueue(dropSubscriptionEvent)
	}
}

func (s *persistentSubscription) onEventAppeared(s2 *client.EventStoreSubscription, evt *client.ResolvedEvent) error {
	s.enqueue(evt)
	return nil
}

func (s *persistentSubscription) onSubscriptionDropped(s2 *client.EventStoreSubscription, dr client.SubscriptionDropReason, err error) error {
	s.enqueueSubscriptionDropNotification(dr, err)
	return nil
}

func (s *persistentSubscription) enqueue(evt *client.ResolvedEvent) {
	s.queue <- evt
	if atomic.CompareAndSwapInt32(&s.isProcessing, 0, 1) {
		go s.processQueue()
	}
}

func (s *persistentSubscription) processQueue() {
	for {
		for len(s.queue) > 0 {
			e := <- s.queue
			if e == dropSubscriptionEvent {
				if s.dropData == nil {
					s.dropData = &dropData{
						reason: client.SubscriptionDropReason_Unknown,
						err:    errors.New("Drop reason not specified"),
					}
				}
				s.dropSubscription(s.dropData.reason, s.dropData.err)
				return
			}
			if s.dropData != nil {
				s.dropSubscription(s.dropData.reason, s.dropData.err)
				return
			}
			err := s.eventAppeared(s, e)
			if err == nil {
				if s.autoAck {
					err = s.subscription.NotifyEventsProcessed([]uuid.UUID{e.OriginalEvent().EventId()})
				}
				if s.settings.VerboseLogging() {
					log.Debugf("Persistent Subscription to %s: processed event (%s, %d, %s @ %s).",
						s.streamId, e.OriginalEvent().EventStreamId(), e.OriginalEvent().EventNumber(),
						e.OriginalEvent().EventType(), e.OriginalEventNumber())
				}
			}
			if err != nil {
				s.dropSubscription(client.SubscriptionDropReason_EventHandlerException, err)
				return
			}
		}
		atomic.CompareAndSwapInt32(&s.isProcessing, 1, 0)
		if len(s.queue) > 0 && atomic.CompareAndSwapInt32(&s.isProcessing, 0, 1) {
			continue
		}
		break
	}
}

func (s *persistentSubscription) dropSubscription(
	reason client.SubscriptionDropReason,
	erro error,
) error {
	if atomic.CompareAndSwapInt32(&s.isDropped, 0, 1) {
		if s.settings.VerboseLogging() {
			log.Debugf("Persistent Subscription to %s: dropping subscription, reason: %s %v.", s.streamId,
				reason, erro)
		}
		if s.subscription != nil {
			if err := s.subscription.Unsubscribe(); err != nil {
				return err
			}
		}
		if s.subscriptionDropped != nil {
			if err := s.subscriptionDropped(s, reason, erro); err != nil {
				return err
			}
		}
		s.stopped.Done()
	}
	return nil
}
