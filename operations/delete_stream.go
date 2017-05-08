package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/models"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type deleteStream struct {
	*baseOperation
	stream          string
	expectedVersion int
	hardDelete      bool
	resultChannel   chan *models.DeleteResult
}

func NewDeleteStream(
	stream string,
	expectedVersion int,
	hardDelete bool,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.DeleteResult,
) *deleteStream {
	return &deleteStream{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:          stream,
		expectedVersion: expectedVersion,
		hardDelete:      hardDelete,
		resultChannel:   resultChannel,
	}
}

func (o *deleteStream) GetRequestCommand() models.Command {
	return models.Command_DeleteStream
}

func (o *deleteStream) GetRequestMessage() proto.Message {
	expectedVersion := int32(o.expectedVersion)
	requireMaster := false
	return &protobuf.DeleteStream{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		RequireMaster:   &requireMaster,
		HardDelete:      &o.hardDelete,
	}
}

func (o *deleteStream) ParseResponse(p *models.Package) {
	if p.Command != models.Command_DeleteStreamCompleted {
		err := o.handleError(p, models.Command_DeleteStreamCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.DeleteStreamCompleted{}
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

func (o *deleteStream) succeed(msg *protobuf.DeleteStreamCompleted) {
	var commitPosition int64 = -1
	var preparePosition int64 = -1
	if msg.CommitPosition != nil {
		commitPosition = *msg.CommitPosition
	}
	if msg.PreparePosition != nil {
		preparePosition = *msg.PreparePosition
	}
	position, err := models.NewPosition(commitPosition, preparePosition)
	o.resultChannel <- models.NewDeleteResult(position, err)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *deleteStream) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewDeleteResult(nil, err)
	close(o.resultChannel)
	o.isCompleted = true
}
