package gesclient

import "fmt"

type Position struct {
	commitPosition  int64
	preparePosition int64
}

func NewPosition(commitPosition int64, preparePosition int64) (*Position, error) {
	if commitPosition < preparePosition {
		return nil, fmt.Errorf("The commit position cannot be less than the prepare position (%d < %d)",
			commitPosition, preparePosition)
	}
	return &Position{commitPosition, preparePosition}, nil
}

func (p *Position) CommitPosition() int64 {
	return p.commitPosition
}

func (p *Position) PreparePosition() int64 {
	return p.preparePosition
}

// Implements fmt.Stringer
func (p *Position) String() string {
	return fmt.Sprintf("%d/%d", p.commitPosition, p.preparePosition)
}

// Implements fmt.GoStringer
func (p *Position) GoString() string {
	return p.String()
}
