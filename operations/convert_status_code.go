package operations

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
)

func convertStatusCode(result protobuf.ReadStreamEventsCompleted_ReadStreamResult) (models.SliceReadStatus, error) {
	switch result {
	case protobuf.ReadStreamEventsCompleted_Success:
		return models.SliceReadStatus_Success, nil
	case protobuf.ReadStreamEventsCompleted_NoStream:
		return models.SliceReadStatus_StreamNotFound, nil
	case protobuf.ReadStreamEventsCompleted_StreamDeleted:
		return models.SliceReadStatus_StreamDeleted, nil
	default:
		return models.SliceReadStatus_Error, fmt.Errorf("Invalid status code: %s", result)
	}
}
