package client

import "fmt"

type Position struct {
	commitPosition  int64
	preparePosition int64
}

var Position_Start = &Position{0, 0}
var Position_End = &Position{-1, -1}

func NewPosition(commitPosition int64, preparePosition int64) *Position {
	if commitPosition < preparePosition {
		panic(fmt.Sprintf("The commit position cannot be less than the prepare position (%d < %d)", commitPosition,
			preparePosition))
	}
	return &Position{commitPosition, preparePosition}
}

func (p *Position) CommitPosition() int64 {
	return p.commitPosition
}

func (p *Position) PreparePosition() int64 {
	return p.preparePosition
}

func (p *Position) Equals(p2 *Position) bool {
	return p.commitPosition == p2.commitPosition && p.preparePosition == p2.preparePosition
}

func (p *Position) GreaterThan(p2 *Position) bool {
	return p.commitPosition > p2.commitPosition ||
		(p.commitPosition == p2.commitPosition && p.preparePosition > p2.preparePosition)
}

func (p *Position) GreaterThanOrEquals(p2 *Position) bool {
	return p.GreaterThan(p2) || p.Equals(p2)
}

// Implements fmt.Stringer
func (p *Position) String() string {
	return fmt.Sprintf("%d/%d", p.commitPosition, p.preparePosition)
}

// Implements fmt.GoStringer
func (p *Position) GoString() string {
	return p.String()
}
