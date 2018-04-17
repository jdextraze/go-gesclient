package projections

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
	StatusUrl                          string //*url.URL // TODO url
	StateUrl                           string //*url.URL
	ResultUrl                          string //*url.URL
	QueryUrl                           string //*url.URL
	EnableCommandUrl                   string //*url.URL
	DisableCommandUrl                  string //*url.URL
	CheckpointStatus                   string
	BufferedEvents                     int
	WritePendingEventsBeforeCheckpoint int
	WritePendingEventsAfterCheckpoint  int
}

type ListResult struct {
	Projections []*ProjectionDetails
}
