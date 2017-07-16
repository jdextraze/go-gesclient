package client

type DeleteResult struct {
	logPosition *Position
}

func NewDeleteResult(logPosition *Position) *DeleteResult {
	return &DeleteResult{
		logPosition: logPosition,
	}
}

func (r *DeleteResult) LogPosition() *Position { return r.logPosition }
