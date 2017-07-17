package client

import (
	"github.com/jdextraze/go-gesclient/messages"
)

type AllEventsSlice struct {
	readDirection ReadDirection
	fromPosition  *Position
	nextPosition  *Position
	events        []*ResolvedEvent
}

func NewAllEventsSlice(
	readDirection ReadDirection,
	fromPosition *Position,
	nextPosition *Position,
	resolvedEvents []*messages.ResolvedEvent,
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
	}
}

func (s *AllEventsSlice) GetReadDirection() ReadDirection { return s.readDirection }

func (s *AllEventsSlice) GetFromPosition() *Position { return s.fromPosition }

func (s *AllEventsSlice) GetNextPosition() *Position { return s.nextPosition }

func (s *AllEventsSlice) GetEvents() []*ResolvedEvent { return s.events }

func (s *AllEventsSlice) IsEndOfStream() bool { return len(s.events) == 0 }
