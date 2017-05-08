package internal

import (
	"net"
	"errors"
)

type NodeEndpoints struct {
	tcpEndpoint *net.TCPAddr
	secureTcpEndpoint *net.TCPAddr
}

func NewNodeEndpoints(
	tcpEndpoint *net.TCPAddr,
	secureTcpEndpoint *net.TCPAddr,
) (*NodeEndpoints, error) {
	if tcpEndpoint == nil && secureTcpEndpoint == nil {
		return nil, errors.New("Both endpoints are nil")
	}
	return &NodeEndpoints{
		tcpEndpoint: tcpEndpoint,
		secureTcpEndpoint: secureTcpEndpoint,
	}, nil
}

func (ne *NodeEndpoints) TcpEndpoint() *net.TCPAddr { return ne.tcpEndpoint }
func (ne *NodeEndpoints) SecureTcpEndpoint() *net.TCPAddr { return ne.secureTcpEndpoint }
