package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/tasks"
)

type readStreamEventsForward struct {
	*baseOperation
	stream         string
	start          int
	max            int
	resolveLinkTos bool
	requireMaster  bool
}

func NewReadStreamEventsForward(
	source *tasks.CompletionSource,
	stream string,
	start int,
	max int,
	resolveLinkTos bool,
	requireMaster bool,
	userCredentials *client.UserCredentials,
) *readStreamEventsForward {
	obj := &readStreamEventsForward{
		stream:         stream,
		start:          start,
		max:            max,
		resolveLinkTos: resolveLinkTos,
		requireMaster:  requireMaster,
	}
	obj.baseOperation = newBaseOperation(client.Command_ReadStreamEventsForward,
		client.Command_ReadStreamEventsForwardCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *readStreamEventsForward) createRequestDto() proto.Message {
	start := int32(o.start)
	max := int32(o.max)
	return &protobuf.ReadStreamEvents{
		EventStreamId:   &o.stream,
		FromEventNumber: &start,
		MaxCount:        &max,
		ResolveLinkTos:  &o.resolveLinkTos,
		RequireMaster:   &o.requireMaster,
	}
}

func (o *readStreamEventsForward) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*protobuf.ReadStreamEventsCompleted)
	switch msg.GetResult() {
	case protobuf.ReadStreamEventsCompleted_Success,
		protobuf.ReadStreamEventsCompleted_StreamDeleted,
		protobuf.ReadStreamEventsCompleted_NoStream:
		o.succeed()
	case protobuf.ReadStreamEventsCompleted_Error:
		o.Fail(client.NewServerError(msg.GetError()))
	case protobuf.ReadStreamEventsCompleted_NotModified:
		o.Fail(client.NewNotModified(o.stream))
	case protobuf.ReadStreamEventsCompleted_AccessDenied:
		o.Fail(client.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result)
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *readStreamEventsForward) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.ReadStreamEventsCompleted)
	status, err := convertStatusCode(msg.GetResult())
	if err != nil {
		return nil, err
	}
	return client.NewStreamEventsSlice(status, o.stream, o.start, client.ReadDirectionForward, msg.GetEvents(),
		int(msg.GetNextEventNumber()), int(msg.GetLastEventNumber()), msg.GetIsEndOfStream()), nil
}

func (o *readStreamEventsForward) createResponse() proto.Message {
	return &protobuf.ReadStreamEventsCompleted{}
}

func (o *readStreamEventsForward) String() string {
	return fmt.Sprintf("ReadStreamEventsForward from stream '%s'", o.stream)
}
