package operations

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
)

func convertStatusCode(result protobuf.ReadStreamEventsCompleted_ReadStreamResult) (client.SliceReadStatus, error) {
	switch result {
	case protobuf.ReadStreamEventsCompleted_Success:
		return client.SliceReadStatus_Success, nil
	case protobuf.ReadStreamEventsCompleted_NoStream:
		return client.SliceReadStatus_StreamNotFound, nil
	case protobuf.ReadStreamEventsCompleted_StreamDeleted:
		return client.SliceReadStatus_StreamDeleted, nil
	default:
		return client.SliceReadStatus_Error, fmt.Errorf("Invalid status code: %s", result)
	}
}
