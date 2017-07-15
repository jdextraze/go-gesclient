package subscriptions

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/tasks"
	"errors"
	"time"
)

type AllCatchUpSubscription struct {
	*catchUpSubscription
	nextReadPosition *models.Position
	lastProcessedPosition *models.Position
	completion *tasks.CompletionSource
}

func NewAllCatchUpSubscription(
	connection models.Connection,
	fromPositionExclusive *models.Position,
	userCredentials *models.UserCredentials,
	eventAppeared models.CatchUpEventAppearedHandler,
	liveProcessingStarted models.LiveProcessingStartedHandler,
	subscriptionDropped models.CatchUpSubscriptionDroppedHandler,
	settings *models.CatchUpSubscriptionSettings,
) *AllCatchUpSubscription {
	var (
		lastProcessedPosition *models.Position
		nextReadPosition *models.Position
	)
	if fromPositionExclusive != nil {
		lastProcessedPosition = fromPositionExclusive
		nextReadPosition = fromPositionExclusive
	} else {
		lastProcessedPosition = models.Position_End
		nextReadPosition = models.Position_Start
	}
	obj := &AllCatchUpSubscription{
		nextReadPosition: nextReadPosition,
		lastProcessedPosition: lastProcessedPosition,
	}
	obj.catchUpSubscription = newCatchUpSubscription(connection, "", userCredentials, eventAppeared,
		liveProcessingStarted, subscriptionDropped, settings, obj.readEventsTillAsync, obj.tryProcess)
	return obj
}

func (s *AllCatchUpSubscription) readEventsTillAsync(
	connection models.Connection,
	resolveLinkTos bool,
	userCredentials *models.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) *tasks.Task {
	s.completion = tasks.NewCompletionSource()
	s.readEventsInternal(connection, resolveLinkTos, userCredentials, lastCommitPosition, lastEventNumber)
	return s.completion.Task()
}

func (s *AllCatchUpSubscription) readEventsInternal(
	connection models.Connection,
	resolveLinkTos bool,
	userCredentials *models.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) {
	task, err := connection.ReadAllEventsForwardAsync(s.nextReadPosition, s.readBatchSize, resolveLinkTos,
		userCredentials)
	if err == nil {
		task.ContinueWith(tasks.ContinueWithCallback(func(t *tasks.Task) error {
			return s.readEventsCallback(t, connection, resolveLinkTos, userCredentials, lastCommitPosition,
				lastEventNumber)
		}))
	} else {
		s.completion.TrySetError(err)
	}
}

func (s *AllCatchUpSubscription) readEventsCallback(
	task *tasks.Task,
	connection models.Connection,
	resolveLinkTos bool,
	userCredentials *models.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) error {
	var err error
	if task.IsFaulted() {
		err = task.Wait()
	} else {
		result := &models.AllEventsSlice{}
		if err = task.Result(result); err == nil {
			if ok, err2 := s.processEvents(lastCommitPosition, result); ok && !s.shouldStop {
				s.readEventsInternal(connection, resolveLinkTos, userCredentials, lastCommitPosition, lastEventNumber)
			} else if err2 != nil {
				err = err2
			} else {
				if s.verbose {
					log.Debugf("Catch-up Subscription to %s: finished reading events, nextReadPosition = %s.",
						s.streamId, s.nextReadPosition)
				}
				res := true
				s.completion.TrySetResult(&res)
			}
		}
	}
	if err != nil {
		s.completion.TrySetError(err)
	}
	return nil
}

func (s *AllCatchUpSubscription) processEvents(
	lastCommitPosition *int64,
	slice *models.AllEventsSlice,
) (bool, error) {
	for _, e := range slice.GetEvents() {
		if e.OriginalPosition() == nil {
			return false, errors.New("Subscription event came up with no OriginalPosition.")
		}
		if err := s.tryProcess(e); err != nil {
			return false, err
		}
	}
	s.nextReadPosition = slice.GetNextPosition()

	var done bool
	if lastCommitPosition == nil {
		done = slice.IsEndOfStream()
	} else {
		done = slice.GetNextPosition().GreaterThanOrEquals(models.NewPosition(*lastCommitPosition, *lastCommitPosition))
	}

	if !done && slice.IsEndOfStream() {
		time.Sleep(time.Millisecond)
	}
	return done, nil
}

func (s *AllCatchUpSubscription) tryProcess(e *models.ResolvedEvent) error {
	processed := false
	if e.OriginalPosition().GreaterThan(s.lastProcessedPosition) {
		if err := s.eventAppeared(s, e); err != nil {
			return err
		}
		s.lastProcessedPosition = e.OriginalPosition()
		processed = true
	}
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: %b event (%s, %d, %s @ %s).", s.streamId, processed,
			e.OriginalEvent().EventStreamId(), e.OriginalEvent().EventNumber(), e.OriginalEvent().EventType(),
			e.OriginalPosition())
	}
	return nil
}
