package client

import (
	"github.com/jdextraze/go-gesclient/guid"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/satori/go.uuid"
	"fmt"
)

type EventData struct {
	eventId  uuid.UUID
	typ      string
	isJson   bool
	data     []byte
	metadata []byte
}

func NewEventData(
	eventId uuid.UUID,
	typ string,
	isJson bool,
	data []byte,
	metadata []byte,
) *EventData {
	return &EventData{eventId, typ, isJson, data, metadata}
}

func (e *EventData) EventId() uuid.UUID { return e.eventId }

func (e *EventData) Type() string { return e.typ }

func (e *EventData) IsJson() bool { return e.isJson }

func (e *EventData) Data() []byte { return e.data }

func (e *EventData) Metadata() []byte { return e.metadata }

func (e *EventData) ToNewEvent() *messages.NewEvent {
	var (
		dataContentType     int32
		metadataContentType int32
	)
	if e.isJson {
		dataContentType = 1
	}
	return &messages.NewEvent{
		EventId:             guid.ToBytes(e.eventId),
		EventType:           &e.typ,
		DataContentType:     &dataContentType,
		MetadataContentType: &metadataContentType,
		Data:                e.data,
		Metadata:            e.metadata,
	}
}

func (e *EventData) String() string {
	return fmt.Sprintf("&{eventId:%s type:%s isJson:%t data:[...] metadata:[...]}", e.eventId, e.typ, e.isJson)
}
