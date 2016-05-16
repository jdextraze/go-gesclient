package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"errors"
)

type StreamEventsSlice struct {
	status          SliceReadStatus
	stream          string
	fromEventNumber int
	readDirection   ReadDirection
	events          []*ResolvedEvent
	nextEventNumber int
	lastEventNumber int
	isEndOfStream   bool
}

func newStreamEventsSlice(
	status SliceReadStatus,
	stream string,
	fromEventNumber int,
	readDirection ReadDirection,
	resolvedEvents []*protobuf.ResolvedIndexedEvent,
	nextEventNumber int,
	lastEventNumber int,
	isEndOfStream bool,
) (*StreamEventsSlice, error) {
	if stream == "" {
		return nil, errors.New("Stream cannot be empty")
	}
	events := make([]*ResolvedEvent, len(resolvedEvents))
	for i, evt := range resolvedEvents {
		events[i] = newResolvedEvent(evt)
	}
	return &StreamEventsSlice{
		status:          status,
		stream:          stream,
		fromEventNumber: fromEventNumber,
		readDirection:   readDirection,
		events:          events,
		nextEventNumber: nextEventNumber,
		lastEventNumber: lastEventNumber,
		isEndOfStream:   isEndOfStream,
	}, nil
}

func (s *StreamEventsSlice) Status() SliceReadStatus { return s.status }

func (s *StreamEventsSlice) Stream() string { return s.stream }

func (s *StreamEventsSlice) FromEventNumber() int { return s.fromEventNumber }

func (s *StreamEventsSlice) ReadDirection() ReadDirection { return s.readDirection }

func (s *StreamEventsSlice) Events() []*ResolvedEvent { return s.events }

func (s *StreamEventsSlice) NextEventNumber() int { return s.nextEventNumber }

func (s *StreamEventsSlice) LastEventNumber() int { return s.lastEventNumber }

func (s *StreamEventsSlice) IsEndOfStream() bool { return s.isEndOfStream }
