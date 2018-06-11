package projections

import "fmt"

type ProjectionDetails struct {
	CoreProcessingTime                 int64
	Version                            int
	Epoch                              int
	EffectiveName                      string
	WritesInProgress                   int
	ReadsInProgress                    int
	PartitionsCached                   int
	Status                             string
	StateReason                        string
	Name                               string
	Mode                               string
	Position                           string
	Progress                           float32
	LastCheckpoint                     string
	EventsProcessedAfterRestart        int
	StatusUrl                          string
	StateUrl                           string
	ResultUrl                          string
	QueryUrl                           string
	EnableCommandUrl                   string
	DisableCommandUrl                  string
	CheckpointStatus                   string
	BufferedEvents                     int
	WritePendingEventsBeforeCheckpoint int
	WritePendingEventsAfterCheckpoint  int
}

func (d *ProjectionDetails) String() string {
	return fmt.Sprintf("Name: %s Status: %s Mode: %s", d.Name, d.Status, d.Mode)
}

type ListResult struct {
	Projections []*ProjectionDetails
}
