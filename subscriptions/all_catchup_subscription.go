package subscriptions

import (
	"errors"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type AllCatchUpSubscription struct {
	*catchUpSubscription
	nextReadPosition      *client.Position
	lastProcessedPosition *client.Position
	completion            *tasks.CompletionSource
}

func NewAllCatchUpSubscription(
	connection client.Connection,
	fromPositionExclusive *client.Position,
	userCredentials *client.UserCredentials,
	eventAppeared client.CatchUpEventAppearedHandler,
	liveProcessingStarted client.LiveProcessingStartedHandler,
	subscriptionDropped client.CatchUpSubscriptionDroppedHandler,
	settings *client.CatchUpSubscriptionSettings,
) *AllCatchUpSubscription {
	var (
		lastProcessedPosition *client.Position
		nextReadPosition      *client.Position
	)
	if fromPositionExclusive != nil {
		lastProcessedPosition = fromPositionExclusive
		nextReadPosition = fromPositionExclusive
	} else {
		lastProcessedPosition = client.Position_End
		nextReadPosition = client.Position_Start
	}
	obj := &AllCatchUpSubscription{
		nextReadPosition:      nextReadPosition,
		lastProcessedPosition: lastProcessedPosition,
	}
	obj.catchUpSubscription = newCatchUpSubscription(connection, "", userCredentials, eventAppeared,
		liveProcessingStarted, subscriptionDropped, settings, obj.readEventsTillAsync, obj.tryProcess)
	return obj
}

func (s *AllCatchUpSubscription) readEventsTillAsync(
	connection client.Connection,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) *tasks.Task {
	s.completion = tasks.NewCompletionSource()
	s.readEventsInternal(connection, resolveLinkTos, userCredentials, lastCommitPosition, lastEventNumber)
	return s.completion.Task()
}

func (s *AllCatchUpSubscription) readEventsInternal(
	connection client.Connection,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) {
	task, err := connection.ReadAllEventsForwardAsync(s.nextReadPosition, s.readBatchSize, resolveLinkTos,
		userCredentials)
	if err == nil {
		task.ContinueWith(func(t *tasks.Task) (interface{}, error) {
			return nil, s.readEventsCallback(t, connection, resolveLinkTos, userCredentials, lastCommitPosition,
				lastEventNumber)
		})
	} else {
		s.completion.TrySetError(err)
	}
}

func (s *AllCatchUpSubscription) readEventsCallback(
	task *tasks.Task,
	connection client.Connection,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) error {
	var err error
	if task.IsFaulted() {
		err = task.Wait()
	} else {
		if err = task.Error(); err == nil {
			result := task.Result().(*client.AllEventsSlice)
			if done, err2 := s.processEvents(lastCommitPosition, result); err2 != nil {
				err = err2
			} else if !done && !s.shouldStop {
				s.readEventsInternal(connection, resolveLinkTos, userCredentials, lastCommitPosition, lastEventNumber)
			} else {
				if s.verbose {
					s.debug("finished reading events, nextReadPosition = %s.", s.nextReadPosition)
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
	slice *client.AllEventsSlice,
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
		done = slice.GetNextPosition().GreaterThanOrEquals(client.NewPosition(*lastCommitPosition, *lastCommitPosition))
	}

	if !done && slice.IsEndOfStream() {
		time.Sleep(time.Millisecond)
	}
	return done, nil
}

func (s *AllCatchUpSubscription) tryProcess(e *client.ResolvedEvent) error {
	processed := false
	if e.OriginalPosition().GreaterThan(s.lastProcessedPosition) {
		if err := s.eventAppeared(s, e); err != nil {
			return err
		}
		s.lastProcessedPosition = e.OriginalPosition()
		processed = true
	}
	if s.verbose {
		s.debug("%t event (%s, %d, %s @ %s).", processed, e.OriginalEvent().EventStreamId(),
			e.OriginalEvent().EventNumber(), e.OriginalEvent().EventType(), e.OriginalPosition())
	}
	return nil
}
