package client

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
