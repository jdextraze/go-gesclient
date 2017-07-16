package internal

import (
	"net"
)

type NodeEndpoints struct {
	tcpEndpoint       net.Addr
	secureTcpEndpoint net.Addr
}

func NewNodeEndpoints(
	tcpEndpoint net.Addr,
	secureTcpEndpoint net.Addr,
) *NodeEndpoints {
	if tcpEndpoint == nil && secureTcpEndpoint == nil {
		panic("Both endpoints are nil")
	}
	return &NodeEndpoints{
		tcpEndpoint:       tcpEndpoint,
		secureTcpEndpoint: secureTcpEndpoint,
	}
}

func (ne *NodeEndpoints) TcpEndpoint() net.Addr       { return ne.tcpEndpoint }
func (ne *NodeEndpoints) SecureTcpEndpoint() net.Addr { return ne.secureTcpEndpoint }
