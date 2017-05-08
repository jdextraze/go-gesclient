package models

import (
	"github.com/jdextraze/go-gesclient/protobuf"
)

type AllEventsSlice struct {
	readDirection ReadDirection
	fromPosition  *Position
	nextPosition  *Position
	events        []*ResolvedEvent
	error         error
}

func NewAllEventsSlice(
	readDirection ReadDirection,
	fromPosition *Position,
	nextPosition *Position,
	resolvedEvents []*protobuf.ResolvedEvent,
	error error,
) *AllEventsSlice {
	events := make([]*ResolvedEvent, len(resolvedEvents))
	for i, evt := range resolvedEvents {
		events[i] = NewResolvedEventFrom(evt)
	}
	return &AllEventsSlice{
		readDirection: readDirection,
		fromPosition:  fromPosition,
		nextPosition:  nextPosition,
		events:        events,
		error:         error,
	}
}

func (s *AllEventsSlice) GetReadDirection() ReadDirection { return s.readDirection }

func (s *AllEventsSlice) GetFromPosition() *Position { return s.fromPosition }

func (s *AllEventsSlice) GetNextPosition() *Position { return s.nextPosition }

func (s *AllEventsSlice) GetEvents() []*ResolvedEvent { return s.events }

func (s *AllEventsSlice) IsEndOfStream() bool { return len(s.events) == 0 }

func (s *AllEventsSlice) Error() error { return s.error }
