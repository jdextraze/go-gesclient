package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/models"
)

type ReadEvent struct {
	*baseOperation
	stream        string
	eventNumber   int
	resolveTos    bool
	resultChannel chan *models.EventReadResult
}

func NewReadEvent(
	stream string,
	eventNumber int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.EventReadResult,
) *ReadEvent {
	return &ReadEvent{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		eventNumber:   eventNumber,
		resolveTos:    resolveTos,
		resultChannel: resultChannel,
	}
}

func (o *ReadEvent) GetRequestCommand() models.Command {
	return models.Command_ReadEvent
}

func (o *ReadEvent) GetRequestMessage() proto.Message {
	eventNumber := int32(o.eventNumber)
	requireMaster := false
	return &protobuf.ReadEvent{
		EventStreamId:  &o.stream,
		EventNumber:    &eventNumber,
		ResolveLinkTos: &o.resolveTos,
		RequireMaster:  &requireMaster,
	}
}

func (o *ReadEvent) ParseResponse(p *models.Package) {
	if p.Command != models.Command_ReadEventCompleted {
		if err := o.handleError(p, models.Command_ReadEventCompleted); err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.ReadEventCompleted{}
	if err := proto.Unmarshal(p.Data, msg); err != nil {
		o.Fail(err)
		return
	}

	switch *msg.Result {
	case protobuf.ReadEventCompleted_Success:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_NotFound:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_StreamDeleted:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_NoStream:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_Error:
		o.Fail(models.NewServerError(msg.GetError()))
	case protobuf.ReadEventCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result))
	}
}

func (o *ReadEvent) succeed(msg *protobuf.ReadEventCompleted) {
	o.resultChannel <- models.NewEventReadResult(
		models.EventReadStatus(msg.GetResult()),
		o.stream,
		o.eventNumber,
		msg.Event,
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *ReadEvent) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewEventReadResult(models.EventReadStatus_Success, "", -1, nil, err)
	close(o.resultChannel)
	o.isCompleted = true
}
