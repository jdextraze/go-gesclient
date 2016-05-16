package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type appendToStreamOperation struct {
	*BaseOperation
	events          []*EventData
	stream          string
	expectedVersion int
	c               chan *WriteResult
}

func newAppendToStreamOperation(
	stream string,
	events []*EventData,
	expectedVersion int,
	c chan *WriteResult,
) *appendToStreamOperation {
	return &appendToStreamOperation{
		BaseOperation: &BaseOperation{
			correlationId: uuid.NewV4(),
		},
		stream:          stream,
		expectedVersion: expectedVersion,
		events:          events,
		c:               c,
	}
}

func (o *appendToStreamOperation) GetRequestCommand() tcpCommand { return tcpCommand_WriteEvents }

func (o *appendToStreamOperation) GetRequestMessage() proto.Message {
	requireMaster := false
	newEvents := make([]*protobuf.NewEvent, len(o.events))
	for i, evt := range o.events {
		newEvents[i] = evt.ToNewEvent()
	}
	expectedVersion := int32(o.expectedVersion)
	return &protobuf.WriteEvents{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		Events:          newEvents,
		RequireMaster:   &requireMaster,
	}
}

func (o *appendToStreamOperation) ParseResponse(cmd tcpCommand, msg proto.Message) {
	writeEventsCompleted := msg.(*protobuf.WriteEventsCompleted)
	var commitPosition int64 = -1
	var preparePosition int64 = -1
	if writeEventsCompleted.CommitPosition != nil {
		commitPosition = *writeEventsCompleted.CommitPosition
	}
	if writeEventsCompleted.PreparePosition != nil {
		preparePosition = *writeEventsCompleted.PreparePosition
	}
	position, _ := NewPosition(commitPosition, preparePosition) // TODO handle error
	o.c <- NewWriteResult(int(*writeEventsCompleted.LastEventNumber), position)
	close(o.c)
	o.isCompleted = true
}

func (o *appendToStreamOperation) SetError(err error) error {
	if o.isCompleted {
		return err
	}
	close(o.c)
	return err
}
