package gesclient

type SliceReadStatus int

const (
	SliceReadStatusSuccess        SliceReadStatus = 0
	SliceReadStatusStreamNotFound SliceReadStatus = 1
	SliceReadStatusStreamDeleted  SliceReadStatus = 2
)
