package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type deleteStreamOperation struct {
	*baseOperation
	stream          string
	expectedVersion int
	hardDelete      bool
	resultChannel   chan *DeleteResult
}

func newDeleteStreamOperation(
	stream string,
	expectedVersion int,
	hardDelete bool,
	userCredentials *UserCredentials,
) *deleteStreamOperation {
	return &deleteStreamOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:          stream,
		expectedVersion: expectedVersion,
		hardDelete:      hardDelete,
		resultChannel:   make(chan *DeleteResult, 1),
	}
}

func (o *deleteStreamOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_DeleteStream
}

func (o *deleteStreamOperation) GetRequestMessage() proto.Message {
	expectedVersion := int32(o.expectedVersion)
	requireMaster := false
	return &protobuf.DeleteStream{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		RequireMaster:   &requireMaster,
		HardDelete:      &o.hardDelete,
	}
}

func (o *deleteStreamOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_DeleteStreamCompleted {
		err := o.handleError(p, tcpCommand_DeleteStreamCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.DeleteStreamCompleted{}
	if err := proto.Unmarshal(p.Payload, msg); err != nil {
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
		o.Fail(WrongExpectedVersion)
	case protobuf.OperationResult_StreamDeleted:
		o.Fail(StreamDeleted)
	case protobuf.OperationResult_InvalidTransaction:
		o.Fail(InvalidTransaction)
	case protobuf.OperationResult_AccessDenied:
		o.Fail(AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected operation result: %v", msg.GetResult()))
	}
}

func (o *deleteStreamOperation) succeed(msg *protobuf.DeleteStreamCompleted) {
	var commitPosition int64 = -1
	var preparePosition int64 = -1
	if msg.CommitPosition != nil {
		commitPosition = *msg.CommitPosition
	}
	if msg.PreparePosition != nil {
		preparePosition = *msg.PreparePosition
	}
	position, err := NewPosition(commitPosition, preparePosition)
	o.resultChannel <- NewDeleteResult(position, err)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *deleteStreamOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- NewDeleteResult(nil, err)
	close(o.resultChannel)
	o.isCompleted = true
}
