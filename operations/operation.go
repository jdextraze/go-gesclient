package operations

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	log "github.com/jdextraze/go-gesclient/logger"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
	"net"
	"reflect"
	"sync/atomic"
)

type CreateRequestDtoHandler func() proto.Message
type InspectResponseHandler func(message proto.Message) (*client.InspectionResult, error)
type TransformResponseHandler func(message proto.Message) (interface{}, error)
type CreateResponseHandler func() proto.Message

type baseOperation struct {
	requestCommand    client.Command
	responseCommand   client.Command
	userCredentials   *client.UserCredentials
	source            *tasks.CompletionSource
	response          proto.Message
	completed         int32
	createRequestDto  CreateRequestDtoHandler
	inspectResponse   InspectResponseHandler
	transformResponse TransformResponseHandler
	createResponse    CreateResponseHandler
}

func newBaseOperation(
	requestCommand client.Command,
	responseCommand client.Command,
	userCredentials *client.UserCredentials,
	source *tasks.CompletionSource,
	createRequestDto CreateRequestDtoHandler,
	inspectResponse InspectResponseHandler,
	transformResponse TransformResponseHandler,
	createResponse CreateResponseHandler,
) *baseOperation {
	if source == nil {
		panic("source is nil")
	}
	return &baseOperation{
		source:            source,
		requestCommand:    requestCommand,
		responseCommand:   responseCommand,
		userCredentials:   userCredentials,
		createRequestDto:  createRequestDto,
		inspectResponse:   inspectResponse,
		transformResponse: transformResponse,
		createResponse:    createResponse,
	}
}

func (o *baseOperation) CreateNetworkPackage(correlationId uuid.UUID) (*client.Package, error) {
	var flags client.TcpFlag
	if o.userCredentials != nil {
		flags = client.FlagsAuthenticated
	}
	data, err := proto.Marshal(o.createRequestDto())
	if err != nil {
		return nil, err
	}
	return client.NewTcpPackage(o.requestCommand, flags, correlationId, data, o.userCredentials), err
}

func (o *baseOperation) InspectPackage(p *client.Package) (result *client.InspectionResult) {
	var err error
	if p.Command() == o.responseCommand {
		o.response = o.createResponse()
		if err = proto.Unmarshal(p.Data(), o.response); err == nil {
			result, err = o.inspectResponse(o.response)
		}
	} else {
		switch p.Command() {
		case client.Command_NotAuthenticated:
			result = o.inspectNotAuthenticated(p)
		case client.Command_BadRequest:
			result = o.inspectBadRequest(p)
		case client.Command_NotHandled:
			result, err = o.inspectNotHandled(p)
		default:
			result, err = o.inspectUnexpectedCommand(p, o.responseCommand)
		}
	}
	if err != nil {
		o.Fail(err)
		result = client.NewInspectionResult(client.InspectionDecision_EndOperation, err.Error(), nil, nil)
	}
	return
}

func (o *baseOperation) succeed() error {
	if atomic.CompareAndSwapInt32(&o.completed, 0, 1) {
		if o.response != nil {
			if result, err := o.transformResponse(o.response); err != nil {
				return err
			} else {
				o.source.SetResult(result)
			}
		} else {
			o.source.SetError(errors.New("No result"))
		}
	}
	return nil
}

func (o *baseOperation) Fail(err error) {
	if atomic.CompareAndSwapInt32(&o.completed, 0, 1) {
		o.source.SetError(err)
	}
}

func (o *baseOperation) inspectNotAuthenticated(p *client.Package) *client.InspectionResult {
	msg := string(p.Data())
	if msg == "" {
		msg = "Authentication error"
	}
	o.Fail(errors.New(msg))
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, "NotAuthenticated", nil, nil)
}

func (o *baseOperation) inspectBadRequest(p *client.Package) *client.InspectionResult {
	msg := string(p.Data())
	if msg == "" {
		msg = "<no message>"
	}
	o.Fail(errors.New(msg))
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, fmt.Sprintf("BadRequest - %s", msg), nil,
		nil)
}

func (o *baseOperation) inspectNotHandled(p *client.Package) (*client.InspectionResult, error) {
	var err error
	dto := &messages.NotHandled{}
	if err := proto.Unmarshal(p.Data(), dto); err != nil {
		return nil, fmt.Errorf("Invalid payload for NotHandled: %v", err)
	}
	switch dto.GetReason() {
	case messages.NotHandled_NotReady:
		return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandled - NotReady", nil, nil), nil
	case messages.NotHandled_TooBusy:
		return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandled - TooBusy", nil, nil), nil
	case messages.NotHandled_NotMaster:
		masterInfo := &messages.NotHandled_MasterInfo{}
		if err = proto.Unmarshal(dto.AdditionalInfo, masterInfo); err != nil {
			break
		}
		var tcpEndpoint *net.TCPAddr
		var secureTcpEndpoint *net.TCPAddr
		tcpEndpoint, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", masterInfo.GetExternalTcpAddress(),
			masterInfo.GetExternalTcpPort()))
		if err != nil {
			break
		}
		secureTcpEndpoint, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d",
			masterInfo.GetExternalSecureTcpAddress(), masterInfo.GetExternalSecureTcpPort()))
		if err != nil {
			break
		}
		return client.NewInspectionResult(client.InspectionDecision_Reconnect, "NotHandled - NotMaster",
			tcpEndpoint, secureTcpEndpoint), nil
	default:
		log.Errorf("Unknown NotHandledReason: %s", dto.Reason)
		return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandled - <unknown>", nil, nil), nil
	}
	return nil, err
}

func (o *baseOperation) inspectUnexpectedCommand(
	p *client.Package,
	expectedCommand client.Command,
) (*client.InspectionResult, error) {
	if p.Command() == expectedCommand {
		return nil, fmt.Errorf("Command should not be %s", p.Command())
	}

	log.Errorf(`Unexpected TcpCommand received.
Expected: %v, Actual: %v, CorrelationId: %v
Operation (%s): %v,
TcpPackage data dump: %v`,
		expectedCommand, p.Command, p.CorrelationId, reflect.TypeOf(o).Name(), o, p.Data,
	)
	o.Fail(fmt.Errorf("Command not expected: %s. Expected: %s.", p.Command(), expectedCommand))
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, fmt.Sprintf("Unexpected command - %s",
		p.Command()), nil, nil), nil
}
