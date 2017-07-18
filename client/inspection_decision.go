package client

type InspectionDecision int

const (
	InspectionDecision_DoNothing InspectionDecision = iota
	InspectionDecision_EndOperation
	InspectionDecision_Retry
	InspectionDecision_Reconnect
	InspectionDecision_Subscribed
)

var inspectionDecisionValues = []string{
	"DoNothing",
	"EndOperation",
	"Retry",
	"Reconnect",
	"Subscribed",
}

func (d InspectionDecision) String() string {
	return inspectionDecisionValues[d]
}
