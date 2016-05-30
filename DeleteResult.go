package gesclient

type DeleteResult struct {
	logPosition *Position
	error       error
}

func NewDeleteResult(logPosition *Position, err error) *DeleteResult {
	return &DeleteResult{
		logPosition: logPosition,
		error:       err,
	}
}

func (r *DeleteResult) LogPosition() *Position { return r.logPosition }

func (r *DeleteResult) Error() error { return r.error }
