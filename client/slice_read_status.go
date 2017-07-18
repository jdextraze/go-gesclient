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
