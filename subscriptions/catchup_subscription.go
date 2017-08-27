package subscriptions

import (
	"errors"
	"github.com/jdextraze/go-gesclient/client"
	log "github.com/jdextraze/go-gesclient/logger"
	"github.com/jdextraze/go-gesclient/tasks"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var dropSubscriptionEvent = client.NewResolvedEvent(nil)

type dropData struct {
	reason client.SubscriptionDropReason
	err    error
}

var nilDropReason *dropData = &dropData{client.SubscriptionDropReason_Unknown, nil}

type ReadEventsTillAsyncHandler func(connection client.Connection, resolveLinkTos bool,
	userCredentials *client.UserCredentials, lastCommitPosition *int64, lastEventNumber *int32) *tasks.Task

type TryProcessHandler func(evt *client.ResolvedEvent) error

type catchUpSubscription struct {
	streamId              string
	connection            client.Connection
	resolveLinkTos        bool
	userCredentials       *client.UserCredentials
	readBatchSize         int
	maxPushQueueSize      int
	eventAppeared         client.CatchUpEventAppearedHandler
	liveProcessingStarted client.LiveProcessingStartedHandler
	subscriptionDropped   client.CatchUpSubscriptionDroppedHandler
	verbose               bool
	liveQueue             chan *client.ResolvedEvent
	subscription          client.EventStoreSubscription
	dropData              *dropData
	allowProcessing       bool
	isProcessing          int32
	shouldStop            bool
	isDropped             int32
	stopped               *sync.WaitGroup
	readEventsTillAsync   ReadEventsTillAsyncHandler
	tryProcess            TryProcessHandler
}

func newCatchUpSubscription(
	connection client.Connection,
	streamId string,
	userCredentials *client.UserCredentials,
	eventAppeared client.CatchUpEventAppearedHandler,
	liveProcessingStarted client.LiveProcessingStartedHandler,
	subscriptionDropped client.CatchUpSubscriptionDroppedHandler,
	settings *client.CatchUpSubscriptionSettings,
	readEventsTillAsync ReadEventsTillAsyncHandler,
	tryProcess TryProcessHandler,
) *catchUpSubscription {
	if connection == nil {
		panic("connection is nil")
	}
	if eventAppeared == nil {
		panic("eventAppeared is nil")
	}
	return &catchUpSubscription{
		connection:            connection,
		streamId:              streamId,
		resolveLinkTos:        settings.ResolveLinkTos(),
		userCredentials:       userCredentials,
		readBatchSize:         settings.ReadBatchSize(),
		maxPushQueueSize:      settings.MaxLiveQueueSize(),
		eventAppeared:         eventAppeared,
		liveProcessingStarted: liveProcessingStarted,
		subscriptionDropped:   subscriptionDropped,
		verbose:               settings.VerboseLogging(),
		liveQueue:             make(chan *client.ResolvedEvent, settings.MaxLiveQueueSize()),
		stopped:               &sync.WaitGroup{},
		readEventsTillAsync:   readEventsTillAsync,
		tryProcess:            tryProcess,
		dropData:              nilDropReason,
	}
}

func (s *catchUpSubscription) IsSubscribedToAll() bool { return s.streamId == "" }

func (s *catchUpSubscription) StreamId() string { return s.streamId }

func (s *catchUpSubscription) Start() *tasks.Task {
	if s.verbose {
		s.debug("starting...")
	}
	return s.runSubscription()
}

func (s *catchUpSubscription) Stop(timeout ...time.Duration) (err error) {
	if len(timeout) > 1 {
		panic("invalid number of arguments")
	}
	if s.verbose {
		s.debug("requesting stop...")
		s.debug("unhooking from connection.Connected.")
	}
	if err = s.connection.Connected().Remove(client.EventHandler(s.onReconnect)); err != nil {
		return
	}
	s.shouldStop = true
	s.enqueueSubscriptionDropNotification(client.SubscriptionDropReason_UserInitiated, nil)
	if len(timeout) == 0 {
		return
	}
	if s.verbose {
		log.Debug("Waiting on subscription to stop")
	}
	go func() {
		<-time.After(timeout[0])
		err = errors.New("Could not stop in time")
		s.stopped.Done()
	}()
	s.stopped.Wait()
	return
}

func (s *catchUpSubscription) onReconnect(evt client.Event) error {
	if s.verbose {
		s.debug("recovering after reconnection.")
		s.debug("unhooking from connection.Connected.")
	}
	if err := s.connection.Connected().Remove(client.EventHandler(s.onReconnect)); err != nil {
		return err
	}
	s.runSubscription()
	return nil
}

func (s *catchUpSubscription) runSubscription() *tasks.Task {
	return tasks.NewStarted(s.loadHistoricalEvents).ContinueWith(func(t *tasks.Task) (interface{}, error) {
		return nil, s.handleErrorOrContinue(t, nil)
	})
}

func (s *catchUpSubscription) loadHistoricalEvents() (interface{}, error) {
	if s.verbose {
		s.debug("running...")
	}

	s.stopped.Add(1)
	s.allowProcessing = true

	if !s.shouldStop {
		if s.verbose {
			s.debug("pulling events...")
		}
		s.readEventsTillAsync(s.connection, s.resolveLinkTos, s.userCredentials, nil, nil).
			ContinueWith(func(t *tasks.Task) (interface{}, error) {
				return nil, s.handleErrorOrContinue(t, s.subscribeToStream)
			})
		return nil, nil
	}
	return nil, s.dropSubscription(client.SubscriptionDropReason_UserInitiated, nil)
}

func (s *catchUpSubscription) subscribeToStream() (err error) {
	if !s.shouldStop {
		if s.verbose {
			s.debug("subscribing...")
		}
		var task *tasks.Task
		if s.streamId == "" {
			task, err = s.connection.SubscribeToAllAsync(s.resolveLinkTos, s.enqueuePushedEvent,
				s.serverSubscriptionDropped, s.userCredentials)
		} else {
			task, err = s.connection.SubscribeToStreamAsync(s.streamId, s.resolveLinkTos, s.enqueuePushedEvent,
				s.serverSubscriptionDropped, s.userCredentials)
		}
		if err != nil {
			return err
		}
		task.ContinueWith(func(t *tasks.Task) (interface{}, error) {
			return nil, s.handleErrorOrContinue(t, func() error {
				if err := t.Error(); err != nil {
					return err
				}
				s.subscription = t.Result().(*VolatileEventStoreSubscription).EventStoreSubscription
				s.readMissedHistoricEvents()
				return nil
			})
		})
		return nil
	}
	return s.dropSubscription(client.SubscriptionDropReason_UserInitiated, nil)
}

func (s *catchUpSubscription) readMissedHistoricEvents() {
	if !s.shouldStop {
		if s.verbose {
			s.debug("pulling events (if left)...")
		}
		lastCommitPosition := s.subscription.LastCommitPosition()
		var lastEventNumber *int32
		l := s.subscription.LastEventNumber()
		if l != nil {
			tmp := *l
			tmp2 := int32(tmp)
			lastEventNumber = &tmp2
		}
		s.readEventsTillAsync(s.connection, s.resolveLinkTos, s.userCredentials, &lastCommitPosition,
			lastEventNumber).ContinueWith(func(t *tasks.Task) (interface{}, error) {
			return nil, s.handleErrorOrContinue(t, s.startLiveProcessing)
		})
		return
	}
	s.dropSubscription(client.SubscriptionDropReason_UserInitiated, nil)
}

func (s *catchUpSubscription) startLiveProcessing() error {
	if s.shouldStop {
		return s.dropSubscription(client.SubscriptionDropReason_UserInitiated, nil)
	}
	if s.verbose {
		s.debug("processing live events...")
	}
	if s.liveProcessingStarted != nil {
		s.liveProcessingStarted(s)
	}
	if s.verbose {
		s.debug("hooking to connection.Connected")
	}
	s.connection.Connected().Add(s.onReconnect)
	s.allowProcessing = true
	s.ensureProcessingPushQueue()
	return nil
}

func (s *catchUpSubscription) enqueueSubscriptionDropNotification(reason client.SubscriptionDropReason, err error) {
	dd := dropData{reason, err}
	if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.dropData)), unsafe.Pointer(nilDropReason), unsafe.Pointer(&dd)) {
		s.liveQueue <- dropSubscriptionEvent
		if s.allowProcessing {
			s.ensureProcessingPushQueue()
		}
	}
}

func (s *catchUpSubscription) handleErrorOrContinue(t *tasks.Task, continuation func() error) error {
	if t.IsFaulted() {
		if err := s.dropSubscription(client.SubscriptionDropReason_CatchUpError, t.Error()); err != nil {
			return err
		}
		return t.Wait()
	} else if continuation != nil {
		return continuation()
	}
	return nil
}

func (s *catchUpSubscription) enqueuePushedEvent(s2 client.EventStoreSubscription, e *client.ResolvedEvent) error {
	if s.verbose {
		s.debug("event appeared (%s, %s, %s @ %s).", e.OriginalStreamId(),
			e.OriginalEventNumber(), e.OriginalEvent().EventType(), e.OriginalPosition())
	}

	if len(s.liveQueue) == cap(s.liveQueue) {
		s.enqueueSubscriptionDropNotification(client.SubscriptionDropReason_ProcessingQueueOverflow, nil)
		s2.Unsubscribe()
		return nil
	}

	s.liveQueue <- e

	if s.allowProcessing {
		s.ensureProcessingPushQueue()
	}
	return nil
}

func (s *catchUpSubscription) serverSubscriptionDropped(
	sub client.EventStoreSubscription,
	reason client.SubscriptionDropReason,
	err error,
) error {
	s.enqueueSubscriptionDropNotification(reason, err)
	return nil
}

func (s *catchUpSubscription) ensureProcessingPushQueue() {
	if atomic.CompareAndSwapInt32(&s.isProcessing, 0, 1) {
		go s.processLiveQueue()
	}
}

func (s *catchUpSubscription) processLiveQueue() {
	for {
		for len(s.liveQueue) > 0 {
			e := <-s.liveQueue
			if e == dropSubscriptionEvent {
				if s.dropData == nilDropReason {
					s.dropData = &dropData{
						reason: client.SubscriptionDropReason_Unknown,
						err:    errors.New("Drop reason not specified"),
					}
				}
				s.dropSubscription(s.dropData.reason, s.dropData.err)
				atomic.CompareAndSwapInt32(&s.isProcessing, 1, 0)
				return
			}

			if err := s.tryProcess(e); err != nil {
				s.dropSubscription(client.SubscriptionDropReason_EventHandlerException, err)
			}
		}
		atomic.CompareAndSwapInt32(&s.isProcessing, 1, 0)
		if len(s.liveQueue) > 0 && atomic.CompareAndSwapInt32(&s.isProcessing, 0, 1) {
			continue
		}
		break
	}
}

func (s *catchUpSubscription) dropSubscription(
	reason client.SubscriptionDropReason,
	erro error,
) error {
	if atomic.CompareAndSwapInt32(&s.isDropped, 0, 1) {
		if s.verbose {
			s.debug("dropping subscription, reason: %s %s.", s.streamId,
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

func (s *catchUpSubscription) debug(format string, args ...interface{}) {
	arguments := make([]interface{}, len(args)+1)
	if s.streamId == "" {
		arguments[0] = "<all>"
	} else {
		arguments[0] = s.streamId
	}
	copy(arguments[1:], args)
	log.Debugf("Catch-up Subscription to %s: "+format, arguments...)
}
