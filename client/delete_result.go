package client

import "fmt"

type DeleteResult struct {
	logPosition *Position
}

func NewDeleteResult(logPosition *Position) *DeleteResult {
	return &DeleteResult{
		logPosition: logPosition,
	}
}

func (r *DeleteResult) LogPosition() *Position { return r.logPosition }

func (r *DeleteResult) String() string {
	return fmt.Sprintf("&{logPosition:%+v}", r.logPosition)
}
