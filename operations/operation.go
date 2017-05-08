package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/models"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"reflect"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("operations")

type baseOperation struct {
	correlationId   uuid.UUID
	isCompleted     bool
	retry           bool
	userCredentials *models.UserCredentials
}

func (o *baseOperation) GetCorrelationId() uuid.UUID { return o.correlationId }

func (o *baseOperation) IsCompleted() bool { return o.isCompleted }

func (o *baseOperation) Retry() bool { return o.retry }

func (o *baseOperation) UserCredentials() *models.UserCredentials { return o.userCredentials }

func (o *baseOperation) handleError(p *models.Package, expectedCommand models.Command) error {
	switch p.Command {
	case models.Command_NotAuthenticated:
		return o.handleNotAuthenticated(p)
	case models.Command_BadRequest:
		return o.handleBadRequest(p)
	case models.Command_NotHandled:
		return o.handleNotHandled(p)
	default:
		return o.handleUnexpectedCommand(p, expectedCommand)
	}
}

func (o *baseOperation) handleNotAuthenticated(p *models.Package) error {
	msg := string(p.Data)
	if msg == "" {
		return models.AuthenticationError
	}
	return errors.New(msg)
}

func (o *baseOperation) handleBadRequest(p *models.Package) error {
	msg := string(p.Data)
	if msg == "" {
		return models.BadRequest
	}
	return errors.New(msg)
}

func (o *baseOperation) handleNotHandled(p *models.Package) (err error) {
	msg := &protobuf.NotHandled{}
	if err := proto.Unmarshal(p.Data, msg); err != nil {
		return fmt.Errorf("Invalid payload for NotHandled: %v", err)
	}
	switch *msg.Reason {
	case protobuf.NotHandled_NotReady:
		o.retry = true
	case protobuf.NotHandled_TooBusy:
		o.retry = true
	case protobuf.NotHandled_NotMaster:
		// TODO support reconnect
		err = errors.New("NotHandled - NotMaster not supported")
	default:
		log.Error("Unknown NotHandledReason: %s", msg.Reason.String())
		o.retry = true
	}
	return
}

func (o *baseOperation) handleUnexpectedCommand(p *models.Package, expectedCommand models.Command) error {
	log.Error(`Unexpected TcpCommand received.
Expected: %v, Actual: %v, CorrelationId: %v
Operation (%s): %v,
TcpPackage data dump: %v`,
		expectedCommand, p.Command, p.CorrelationId, reflect.TypeOf(o).Name(), o, p.Data,
	)
	return fmt.Errorf("Command not expected. Expected: %v, Actual: %v", expectedCommand, p.Command)
}
