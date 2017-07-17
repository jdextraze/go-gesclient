package client

import (
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/satori/go.uuid"
	"time"
)

type RecordedEvent struct {
	eventStreamId string
	eventId       uuid.UUID
	eventNumber   int
	eventType     string
	data          []byte
	metadata      []byte
	isJson        bool
	created       time.Time
	createdEpoch  time.Time
}

func newRecordedEvent(evt *messages.EventRecord) *RecordedEvent {
	return &RecordedEvent{
		eventStreamId: evt.GetEventStreamId(),
		eventId:       uuid.FromBytesOrNil(evt.GetEventId()),
		eventNumber:   int(evt.GetEventNumber()),
		eventType:     evt.GetEventType(),
		data:          evt.GetData(),
		metadata:      evt.GetMetadata(),
		isJson:        evt.GetDataContentType() == 1,
		created:       timeFromTicks(evt.GetCreated()),
		createdEpoch:  timeFromUnixMilliseconds(evt.GetCreatedEpoch()),
	}
}

func (e *RecordedEvent) EventStreamId() string { return e.eventStreamId }

func (e *RecordedEvent) EventId() uuid.UUID { return e.eventId }

func (e *RecordedEvent) EventNumber() int { return e.eventNumber }

func (e *RecordedEvent) EventType() string { return e.eventType }

func (e *RecordedEvent) Data() []byte { return e.data }

func (e *RecordedEvent) Metadata() []byte { return e.metadata }

func (e *RecordedEvent) IsJson() bool { return e.isJson }

func (e *RecordedEvent) Created() time.Time { return e.created }

func (e *RecordedEvent) CreatedEpoch() time.Time { return e.createdEpoch }
