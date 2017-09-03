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

var eventReadStatuses = map[int]string{
	0: "Success",
	1: "NotFound",
	2: "NoStream",
	3: "StreamDeleted",
	4: "Error",
	5: "AccessDenied",
}

func (s EventReadStatus) String() string {
	return eventReadStatuses[int(s)]
}
