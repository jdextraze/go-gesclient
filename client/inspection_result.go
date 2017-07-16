package client

import "net"

type InspectionResult struct {
	decision          InspectionDecision
	description       string
	tcpEndpoint       net.Addr
	secureTcpEndpoint net.Addr
}

func NewInspectionResult(
	decision InspectionDecision,
	description string,
	tcpEndpoint net.Addr,
	secureTcpEndpoint net.Addr,
) *InspectionResult {
	if decision == InspectionDecision_Reconnect {
		if tcpEndpoint == nil {
			panic("tcpEndpoint should not be nil")
		}
	} else {
		if tcpEndpoint != nil {
			panic("tcpEndpoint should be nil")
		}
	}
	return &InspectionResult{
		decision:          decision,
		description:       description,
		tcpEndpoint:       tcpEndpoint,
		secureTcpEndpoint: secureTcpEndpoint,
	}
}

func (r *InspectionResult) Decision() InspectionDecision { return r.decision }

func (r *InspectionResult) Description() string { return r.description }

func (r *InspectionResult) TcpEndpoint() net.Addr { return r.tcpEndpoint }

func (r *InspectionResult) SecureTcpEndpoint() net.Addr { return r.secureTcpEndpoint }
