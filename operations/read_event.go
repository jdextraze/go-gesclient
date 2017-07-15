package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/tasks"
)

type ReadEvent struct {
	*baseOperation
	stream        string
	eventNumber   int
	resolveTos    bool
}

func NewReadEvent(
	source *tasks.CompletionSource,
	stream string,
	eventNumber int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
) *ReadEvent {
	obj := &ReadEvent{
		stream:        stream,
		eventNumber:   eventNumber,
		resolveTos:    resolveTos,
	}
	obj.baseOperation = newBaseOperation(models.Command_ReadEvent, models.Command_ReadEventCompleted, userCredentials,
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

func (o *ReadEvent) inspectResponse(message proto.Message) (*models.InspectionResult, error) {
	msg := message.(*protobuf.ReadEventCompleted)
	switch msg.GetResult() {
	case protobuf.ReadEventCompleted_Success, protobuf.ReadEventCompleted_NotFound,
		protobuf.ReadEventCompleted_StreamDeleted, protobuf.ReadEventCompleted_NoStream:
		if err := o.succeed(); err != nil {
			return nil, err
		}
	case protobuf.ReadEventCompleted_Error:
		o.Fail(models.NewServerError(msg.GetError()))
	case protobuf.ReadEventCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result)
	}
	return models.NewInspectionResult(models.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *ReadEvent) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.ReadEventCompleted)
	status, err := o.convert(msg.GetResult())
	if err != nil {
		return nil, err
	}
	return models.NewEventReadResult(status, o.stream, o.eventNumber, msg.Event), nil
}

func (o *ReadEvent) convert(result protobuf.ReadEventCompleted_ReadEventResult) (models.EventReadStatus, error) {
	switch result {
	case protobuf.ReadEventCompleted_Success:
		return models.EventReadStatus_Success, nil
	case protobuf.ReadEventCompleted_NotFound:
		return models.EventReadStatus_NotFound, nil
	case protobuf.ReadEventCompleted_NoStream:
		return models.EventReadStatus_NoStream, nil
	case protobuf.ReadEventCompleted_StreamDeleted:
		return models.EventReadStatus_StreamDeleted, nil
	default:
		return models.EventReadStatus_Error, fmt.Errorf("Unexpected ReadEventResult: %s", result)
	}
}

func (o *ReadEvent) createResponse() proto.Message {
	return &protobuf.ReadEventCompleted{}
}

func (o *ReadEvent) String() string {
	return fmt.Sprintf("ReadEvent from stream '%s'", o.stream)
}
