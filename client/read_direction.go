package client

type ReadDirection int

const (
	ReadDirection_Forward ReadDirection = iota
	ReadDirection_Backward
)

var readDirection_names = []string{
	"Forward",
	"Backward",
}

func (x ReadDirection) String() string {
	return readDirection_names[x]
}
