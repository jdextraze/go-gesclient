package client

type EventReadStatus int

const (
	EventReadStatus_Success       EventReadStatus = 0
	EventReadStatus_NotFound      EventReadStatus = 1
	EventReadStatus_NoStream      EventReadStatus = 2
	EventReadStatus_StreamDeleted EventReadStatus = 3
	EventReadStatus_Error         EventReadStatus = 4
	EventReadStatus_AccessDenied  EventReadStatus = 5
)
