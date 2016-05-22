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
	userCredentials *UserCredentials,
) *appendToStreamOperation {
	return &appendToStreamOperation{
		BaseOperation: &BaseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
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

func (o *appendToStreamOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_WriteEventsCompleted {
		err := o.HandleError(p, tcpCommand_WriteEventsCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}
	msg := &protobuf.WriteEventsCompleted{}
	err := proto.Unmarshal(p.Payload, msg)
	var commitPosition int64 = -1
	var preparePosition int64 = -1
	if msg.CommitPosition != nil {
		commitPosition = *msg.CommitPosition
	}
	if msg.PreparePosition != nil {
		preparePosition = *msg.PreparePosition
	}
	position, err := NewPosition(commitPosition, preparePosition)
	o.c <- NewWriteResult(int(*msg.LastEventNumber), position, err)
	close(o.c)
	o.isCompleted = true
}

func (o *appendToStreamOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.c <- NewWriteResult(0, nil, err)
	close(o.c)
	o.isCompleted = true
}
