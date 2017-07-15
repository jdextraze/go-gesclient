package subscriptions

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/tasks"
	"sync"
	"time"
	"errors"
	"sync/atomic"
)

var dropSubscriptionEvent = models.NewResolvedEvent(nil)

type dropData struct {
	reason models.SubscriptionDropReason
	err    error
}

type ReadEventsTillAsyncHandler func(connection models.Connection, resolveLinkTos bool,
	userCredentials *models.UserCredentials, lastCommitPosition *int64, lastEventNumber *int32) *tasks.Task

type TryProcessHandler func(evt *models.ResolvedEvent) error

type catchUpSubscription struct {
	streamId              string
	connection            models.Connection
	resolveLinkTos        bool
	userCredentials       *models.UserCredentials
	readBatchSize         int
	maxPushQueueSize      int
	eventAppeared         models.CatchUpEventAppearedHandler
	liveProcessingStarted models.LiveProcessingStartedHandler
	subscriptionDropped   models.CatchUpSubscriptionDroppedHandler
	verbose               bool
	liveQueue             chan *models.ResolvedEvent
	subscription          *models.EventStoreSubscription
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
	connection models.Connection,
	streamId string,
	userCredentials *models.UserCredentials,
	eventAppeared models.CatchUpEventAppearedHandler,
	liveProcessingStarted models.LiveProcessingStartedHandler,
	subscriptionDropped models.CatchUpSubscriptionDroppedHandler,
	settings *models.CatchUpSubscriptionSettings,
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
		connection: connection,
		streamId: streamId,
		resolveLinkTos: settings.ResolveLinkTos(),
		userCredentials: userCredentials,
		readBatchSize: settings.ReadBatchSize(),
		maxPushQueueSize: settings.MaxLiveQueueSize(),
		eventAppeared: eventAppeared,
		liveProcessingStarted: liveProcessingStarted,
		subscriptionDropped: subscriptionDropped,
		verbose: settings.VerboseLogging(),
		liveQueue: make(chan *models.ResolvedEvent, settings.MaxLiveQueueSize()),
		stopped: &sync.WaitGroup{},
		readEventsTillAsync: readEventsTillAsync,
		tryProcess: tryProcess,
	}
}

func (s *catchUpSubscription) IsSubscribedToAll() bool { return s.streamId == "" }

func (s *catchUpSubscription) StreamId() string { return s.streamId }

func (s *catchUpSubscription) Start() *tasks.Task {
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: starting...", s.streamId)
	}
	return s.runSubscription()
}

func (s *catchUpSubscription) Stop(timeout ...time.Duration) (err error) {
	if len(timeout) > 1 {
		panic("invalid number of arguments")
	}
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: requesting stop...", s.streamId)
		log.Debugf("Catch-up Subscription to %s: unhooking from connection.Connected.", s.streamId)
	}
	if err = s.connection.Connected().Remove(models.EventHandler(s.onReconnect)); err != nil {
		return
	}
	s.shouldStop = true
	s.enqueueSubscriptionDropNotification(models.SubscriptionDropReason_UserInitiated, nil)
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

func (s *catchUpSubscription) onReconnect(evt models.Event) error {
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: recovering after reconnection.", s.streamId)
		log.Debugf("Catch-up Subscription to %s: unhooking from connection.Connected.", s.streamId)
	}
	if err := s.connection.Connected().Remove(models.EventHandler(s.onReconnect)); err != nil {
		return err
	}
	s.runSubscription()
	return nil
}

func (s *catchUpSubscription) runSubscription() *tasks.Task {
	return tasks.NewStarted(tasks.TaskCallback(s.loadHistoricalEvents)).
		ContinueWith(tasks.ContinueWithCallback(func(t *tasks.Task) error {
		return s.handleErrorOrContinue(t, nil)
	}))
}

func (s *catchUpSubscription) loadHistoricalEvents() (interface{}, error) {
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: running...", s.streamId)
	}

	s.stopped.Add(1)
	s.allowProcessing = true

	if !s.shouldStop {
		if s.verbose {
			log.Debugf("Catch-up Subscription to %s: pulling events...", s.streamId)
		}
		s.readEventsTillAsync(s.connection, s.resolveLinkTos, s.userCredentials, nil, nil).
			ContinueWith(tasks.ContinueWithCallback(func(t *tasks.Task) error {
			return s.handleErrorOrContinue(t, s.subscribeToStream)
		}))
		return nil, nil
	}
	return nil, s.dropSubscription(models.SubscriptionDropReason_UserInitiated, nil)
}

func (s *catchUpSubscription) subscribeToStream() (err error) {
	if !s.shouldStop {
		if s.verbose {
			log.Debugf("Catch-up Subscription to %s: subscribing...", s.streamId)
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
		task.ContinueWith(tasks.ContinueWithCallback(func(t *tasks.Task) error {
			return s.handleErrorOrContinue(t, func() error {
				s.subscription = &models.EventStoreSubscription{}
				if err := t.Result(s.subscription); err != nil {
					return err
				}
				s.readMissedHistoricEvents()
				return nil
			})
		}))
	}
	return s.dropSubscription(models.SubscriptionDropReason_UserInitiated, nil)
}

func (s *catchUpSubscription) readMissedHistoricEvents() {
	if !s.shouldStop {
		if s.verbose {
			log.Debugf("Catch-up Subscription to %s: pulling events (if left)...", s.streamId)
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
			lastEventNumber).ContinueWith(tasks.ContinueWithCallback(func(t *tasks.Task) error {
			return s.handleErrorOrContinue(t, s.startLiveProcessing)
		}))
	} else {
		s.dropSubscription(models.SubscriptionDropReason_UserInitiated, nil)
	}
}

func (s *catchUpSubscription) startLiveProcessing() error {
	if s.shouldStop {
		return s.dropSubscription(models.SubscriptionDropReason_UserInitiated, nil)
	}
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: processing live events...", s.streamId)
	}
	if s.liveProcessingStarted != nil {
		s.liveProcessingStarted(s)
	}
	if s.verbose {
		log.Debugf("Catch-up Subscription to {0}: hooking to connection.Connected", s.streamId)
	}
	s.connection.Connected().Add(s.onReconnect)
	s.allowProcessing = true
	s.ensureProcessingPushQueue()
	return nil
}

func (s *catchUpSubscription) enqueueSubscriptionDropNotification(
	reason models.SubscriptionDropReason,
	err error,
) {}

func (s *catchUpSubscription) handleErrorOrContinue(t *tasks.Task, continuation func() error) error {
	if t.IsFaulted() {
		if err := s.dropSubscription(models.SubscriptionDropReason_CatchUpError, t.Error()); err != nil {
			return err
		}
		return t.Wait()
	} else if continuation != nil {
		return continuation()
	}
	return nil
}

func (s *catchUpSubscription) enqueuePushedEvent(s2 *models.EventStoreSubscription, e *models.ResolvedEvent) error {
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: event appeared (%s, %s, %s @ %s).", s.streamId, e.OriginalStreamId(),
			e.OriginalEventNumber(), e.OriginalEvent().EventType(), e.OriginalPosition())
	}

	if len(s.liveQueue) == cap(s.liveQueue) {
		s.enqueueSubscriptionDropNotification(models.SubscriptionDropReason_ProcessingQueueOverflow, nil)
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
	sub *models.EventStoreSubscription,
	reason models.SubscriptionDropReason,
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
		e := <- s.liveQueue
		if e == dropSubscriptionEvent {
			if s.dropData == nil {
				s.dropData = &dropData{
					reason: models.SubscriptionDropReason_Unknown,
					err: errors.New("Drop reason not specified"),
				}
				s.dropSubscription(s.dropData.reason, s.dropData.err)
				atomic.CompareAndSwapInt32(&s.isProcessing, 1, 0)
				return
			}
		}

		if err := s.tryProcess(e); err != nil {
			s.dropSubscription(models.SubscriptionDropReason_EventHandlerException, err)
		}

		atomic.CompareAndSwapInt32(&s.isProcessing, 1, 0)
		if len(s.liveQueue) > 0 && atomic.CompareAndSwapInt32(&s.isProcessing, 0, 1) {
			continue
		}
		break
	}
}

func (s *catchUpSubscription) dropSubscription(
	reason models.SubscriptionDropReason,
	erro error,
) error {
	if atomic.CompareAndSwapInt32(&s.isDropped, 0, 1) {
		if s.verbose {
			log.Debugf("Catch-up Subscription to %s: dropping subscription, reason: %s %s.", s.streamId,
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
