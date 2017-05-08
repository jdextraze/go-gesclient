package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/models"
)

type appendToStream struct {
	*baseOperation
	events          []*models.EventData
	stream          string
	expectedVersion int
	resultChannel   chan *models.WriteResult
}

func NewAppendToStream(
	stream string,
	events []*models.EventData,
	expectedVersion int,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.WriteResult,
) *appendToStream {
	return &appendToStream{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:          stream,
		expectedVersion: expectedVersion,
		events:          events,
		resultChannel:   resultChannel,
	}
}

func (o *appendToStream) GetRequestCommand() models.Command {
	return models.Command_WriteEvents
}

func (o *appendToStream) GetRequestMessage() proto.Message {
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

func (o *appendToStream) ParseResponse(p *models.Package) {
	if p.Command != models.Command_WriteEventsCompleted {
		err := o.handleError(p, models.Command_WriteEventsCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.WriteEventsCompleted{}
	if err := proto.Unmarshal(p.Data, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.OperationResult_Success:
		o.succeed(msg)
	case protobuf.OperationResult_PrepareTimeout,
		protobuf.OperationResult_ForwardTimeout,
		protobuf.OperationResult_CommitTimeout:
		o.retry = true
	case protobuf.OperationResult_WrongExpectedVersion:
		o.Fail(models.WrongExpectedVersion)
	case protobuf.OperationResult_StreamDeleted:
		o.Fail(models.StreamDeleted)
	case protobuf.OperationResult_InvalidTransaction:
		o.Fail(models.InvalidTransaction)
	case protobuf.OperationResult_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected Operation result: %v", msg.GetResult()))
	}
}

func (o *appendToStream) succeed(msg *protobuf.WriteEventsCompleted) {
	var commitPosition int64 = -1
	var preparePosition int64 = -1
	if msg.CommitPosition != nil {
		commitPosition = *msg.CommitPosition
	}
	if msg.PreparePosition != nil {
		preparePosition = *msg.PreparePosition
	}
	position, err := models.NewPosition(commitPosition, preparePosition)
	o.resultChannel <- models.NewWriteResult(int(*msg.LastEventNumber), position, err)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *appendToStream) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewWriteResult(0, nil, err)
	close(o.resultChannel)
	o.isCompleted = true
}
