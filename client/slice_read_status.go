package client

type SliceReadStatus int

const (
	SliceReadStatus_Success        SliceReadStatus = 0
	SliceReadStatus_StreamNotFound SliceReadStatus = 1
	SliceReadStatus_StreamDeleted  SliceReadStatus = 2
	SliceReadStatus_NotModified    SliceReadStatus = 3
	SliceReadStatus_Error          SliceReadStatus = 4
	SliceReadStatus_AccessDenied   SliceReadStatus = 5
)

var sliceReadStatuses = map[int]string{
	0: "Success",
	1: "StreamNotFound",
	2: "StreamDeleted",
	3: "NotModified",
	4: "Error",
	5: "AccessDenied",
}

func (s SliceReadStatus) String() string {
	return sliceReadStatuses[int(s)]
}
