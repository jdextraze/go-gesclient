package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/models"
)

type readAllEventsForward struct {
	*baseOperation
	pos           *models.Position
	max           int
	resolveTos    bool
	resultChannel chan *models.AllEventsSlice
}

func NewReadAllEventsForward(
	pos *models.Position,
	max int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.AllEventsSlice,
) *readAllEventsForward {
	return &readAllEventsForward{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		pos:           pos,
		max:           max,
		resolveTos:    resolveTos,
		resultChannel: resultChannel,
	}
}

func (o *readAllEventsForward) GetRequestCommand() models.Command {
	return models.Command_ReadAllEventsForward
}

func (o *readAllEventsForward) GetRequestMessage() proto.Message {
	commitPos := o.pos.CommitPosition()
	preparePos := o.pos.PreparePosition()
	no := false
	max := int32(o.max)
	return &protobuf.ReadAllEvents{
		CommitPosition:  &commitPos,
		PreparePosition: &preparePos,
		MaxCount:        &max,
		ResolveLinkTos:  &no,
		RequireMaster:   &no,
	}
}

func (o *readAllEventsForward) ParseResponse(p *models.Package) {
	if p.Command != models.Command_ReadAllEventsForwardCompleted {
		err := o.handleError(p, models.Command_ReadAllEventsForwardCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.ReadAllEventsCompleted{}
	err := proto.Unmarshal(p.Data, msg)
	if err != nil {
		o.Fail(err)
		return
	}

	if msg.Result == nil {
		o.succeed(msg)
		return
	}

	switch *msg.Result {
	case protobuf.ReadAllEventsCompleted_Success:
		o.succeed(msg)
	case protobuf.ReadAllEventsCompleted_Error:
		o.Fail(models.NewServerError(msg.GetError()))
	case protobuf.ReadAllEventsCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected ReadAllResult: %v", *msg.Result))
	}
}

func (o *readAllEventsForward) succeed(msg *protobuf.ReadAllEventsCompleted) {
	fromPosition, _ := models.NewPosition(*msg.CommitPosition, *msg.PreparePosition)
	nextPosition, _ := models.NewPosition(*msg.NextCommitPosition, *msg.NextPreparePosition)
	o.resultChannel <- models.NewAllEventsSlice(
		models.ReadDirectionForward,
		fromPosition,
		nextPosition,
		msg.Events,
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *readAllEventsForward) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewAllEventsSlice(
		models.ReadDirectionForward,
		nil,
		nil,
		[]*protobuf.ResolvedEvent{},
		err,
	)
	close(o.resultChannel)
	o.isCompleted = true
}
