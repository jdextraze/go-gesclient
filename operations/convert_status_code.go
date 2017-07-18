package operations

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
)

func convertStatusCode(result messages.ReadStreamEventsCompleted_ReadStreamResult) (client.SliceReadStatus, error) {
	switch result {
	case messages.ReadStreamEventsCompleted_Success:
		return client.SliceReadStatus_Success, nil
	case messages.ReadStreamEventsCompleted_NoStream:
		return client.SliceReadStatus_StreamNotFound, nil
	case messages.ReadStreamEventsCompleted_StreamDeleted:
		return client.SliceReadStatus_StreamDeleted, nil
	default:
		return client.SliceReadStatus_Error, fmt.Errorf("Invalid status code: %s", result)
	}
}
