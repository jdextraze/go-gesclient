package internal

import (
	"fmt"
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

func (e *NodeEndpoints) TcpEndpoint() net.Addr { return e.tcpEndpoint }

func (e *NodeEndpoints) SecureTcpEndpoint() net.Addr { return e.secureTcpEndpoint }

func (e *NodeEndpoints) String() string {
	normal := "n/a"
	secure := "n/a"
	if e.tcpEndpoint != nil {
		normal = e.tcpEndpoint.String()
	}
	if e.secureTcpEndpoint != nil {
		secure = e.secureTcpEndpoint.String()
	}
	return fmt.Sprintf("[%s, %s]", normal, secure)
}
