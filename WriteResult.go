package gesclient

type WriteResult struct {
	nextExpectedVersion int
	logPosition         *Position
	error               error
}

func NewWriteResult(nextExpectedVersion int, logPosition *Position, err error) *WriteResult {
	return &WriteResult{nextExpectedVersion, logPosition, err}
}

func (r *WriteResult) NextExpectedVersion() int {
	return r.nextExpectedVersion
}

func (r *WriteResult) LogPosition() *Position {
	return r.logPosition
}

func (r *WriteResult) Error() error {
	return r.error
}
