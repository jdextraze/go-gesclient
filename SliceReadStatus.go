package gesclient

type SliceReadStatus int

const (
	SliceReadStatusSuccess        SliceReadStatus = 0
	SliceReadStatusStreamNotFound SliceReadStatus = 1
	SliceReadStatusStreamDeleted  SliceReadStatus = 2
	SliceReadStatusNotModified    SliceReadStatus = 3
	SliceReadStatusError          SliceReadStatus = 4
	SliceReadStatusAccessDenied   SliceReadStatus = 5
)
