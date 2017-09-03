package client

import "fmt"

type WriteResult struct {
	nextExpectedVersion int
	logPosition         *Position
}

func NewWriteResult(nextExpectedVersion int, logPosition *Position) *WriteResult {
	return &WriteResult{
		nextExpectedVersion: nextExpectedVersion,
		logPosition:         logPosition,
	}
}

func (r *WriteResult) NextExpectedVersion() int {
	return r.nextExpectedVersion
}

func (r *WriteResult) LogPosition() *Position {
	return r.logPosition
}

func (r *WriteResult) String() string {
	return fmt.Sprintf("WriteResult{nextExpectedVersion: %d logPosition: %s}", r.nextExpectedVersion, r.logPosition)
}
