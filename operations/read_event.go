package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/tasks"
)

type ReadEvent struct {
	*baseOperation
	stream      string
	eventNumber int
	resolveTos  bool
}

func NewReadEvent(
	source *tasks.CompletionSource,
	stream string,
	eventNumber int,
	resolveTos bool,
	userCredentials *client.UserCredentials,
) *ReadEvent {
	obj := &ReadEvent{
		stream:      stream,
		eventNumber: eventNumber,
		resolveTos:  resolveTos,
	}
	obj.baseOperation = newBaseOperation(client.Command_ReadEvent, client.Command_ReadEventCompleted, userCredentials,
		source, obj.createRequestDto, obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *ReadEvent) createRequestDto() proto.Message {
	eventNumber := int32(o.eventNumber)
	requireMaster := false
	return &protobuf.ReadEvent{
		EventStreamId:  &o.stream,
		EventNumber:    &eventNumber,
		ResolveLinkTos: &o.resolveTos,
		RequireMaster:  &requireMaster,
	}
}

func (o *ReadEvent) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*protobuf.ReadEventCompleted)
	switch msg.GetResult() {
	case protobuf.ReadEventCompleted_Success, protobuf.ReadEventCompleted_NotFound,
		protobuf.ReadEventCompleted_StreamDeleted, protobuf.ReadEventCompleted_NoStream:
		if err := o.succeed(); err != nil {
			return nil, err
		}
	case protobuf.ReadEventCompleted_Error:
		o.Fail(client.NewServerError(msg.GetError()))
	case protobuf.ReadEventCompleted_AccessDenied:
		o.Fail(client.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result)
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *ReadEvent) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.ReadEventCompleted)
	status, err := o.convert(msg.GetResult())
	if err != nil {
		return nil, err
	}
	return client.NewEventReadResult(status, o.stream, o.eventNumber, msg.Event), nil
}

func (o *ReadEvent) convert(result protobuf.ReadEventCompleted_ReadEventResult) (client.EventReadStatus, error) {
	switch result {
	case protobuf.ReadEventCompleted_Success:
		return client.EventReadStatus_Success, nil
	case protobuf.ReadEventCompleted_NotFound:
		return client.EventReadStatus_NotFound, nil
	case protobuf.ReadEventCompleted_NoStream:
		return client.EventReadStatus_NoStream, nil
	case protobuf.ReadEventCompleted_StreamDeleted:
		return client.EventReadStatus_StreamDeleted, nil
	default:
		return client.EventReadStatus_Error, fmt.Errorf("Unexpected ReadEventResult: %s", result)
	}
}

func (o *ReadEvent) createResponse() proto.Message {
	return &protobuf.ReadEventCompleted{}
}

func (o *ReadEvent) String() string {
	return fmt.Sprintf("ReadEvent from stream '%s'", o.stream)
}
