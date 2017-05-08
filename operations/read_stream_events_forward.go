package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/models"
)

type readStreamEventsForward struct {
	*baseOperation
	stream        string
	start         int
	max           int
	resultChannel chan *models.StreamEventsSlice
}

func NewReadStreamEventsForward(
	stream string,
	start int,
	max int,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.StreamEventsSlice,
) *readStreamEventsForward {
	return &readStreamEventsForward{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		start:         start,
		max:           max,
		resultChannel: resultChannel,
	}
}

func (o *readStreamEventsForward) GetRequestCommand() models.Command {
	return models.Command_ReadStreamEventsForward
}

func (o *readStreamEventsForward) GetRequestMessage() proto.Message {
	no := false
	start := int32(o.start)
	max := int32(o.max)
	return &protobuf.ReadStreamEvents{
		EventStreamId:   &o.stream,
		FromEventNumber: &start,
		MaxCount:        &max,
		ResolveLinkTos:  &no,
		RequireMaster:   &no,
	}
}

func (o *readStreamEventsForward) ParseResponse(p *models.Package) {
	if p.Command != models.Command_ReadStreamEventsForwardCompleted {
		err := o.handleError(p, models.Command_ReadStreamEventsForwardCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.ReadStreamEventsCompleted{}
	err := proto.Unmarshal(p.Data, msg)
	if err != nil {
		o.Fail(err)
		return
	}

	switch *msg.Result {
	case protobuf.ReadStreamEventsCompleted_Success:
		o.succeed(msg)
	case protobuf.ReadStreamEventsCompleted_StreamDeleted:
		o.succeed(msg)
	case protobuf.ReadStreamEventsCompleted_NoStream:
		o.succeed(msg)
	case protobuf.ReadStreamEventsCompleted_Error:
		o.Fail(models.NewServerError(msg.GetError()))
	case protobuf.ReadStreamEventsCompleted_NotModified:
		o.Fail(models.NewNotModified(o.stream))
	case protobuf.ReadStreamEventsCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result))
	}
}

func (o *readStreamEventsForward) succeed(msg *protobuf.ReadStreamEventsCompleted) {
	o.resultChannel <- models.NewStreamEventsSlice(
		models.SliceReadStatus(msg.GetResult()),
		o.stream,
		o.start,
		models.ReadDirectionForward,
		msg.Events,
		int(msg.GetNextEventNumber()),
		int(msg.GetLastEventNumber()),
		msg.GetIsEndOfStream(),
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *readStreamEventsForward) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewStreamEventsSlice(
		models.SliceReadStatus_Error,
		o.stream,
		o.start,
		models.ReadDirectionForward,
		[]*protobuf.ResolvedIndexedEvent{},
		0,
		0,
		false,
		err,
	)
	close(o.resultChannel)
	o.isCompleted = true
}
