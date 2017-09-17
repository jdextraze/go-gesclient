package client

import (
	"fmt"
	"net"
)

type GossipSeed struct {
	ipEndpoint *net.TCPAddr
	hostHeader string
}

func NewGossipSeed(ipEndpoint *net.TCPAddr, hostHeader string) *GossipSeed {
	return &GossipSeed{
		ipEndpoint: ipEndpoint,
		hostHeader: hostHeader,
	}
}

func (gs *GossipSeed) IpEndpoint() *net.TCPAddr {
	return gs.ipEndpoint
}

func (gs *GossipSeed) HostHeader() string {
	return gs.hostHeader
}

func (gs *GossipSeed) String() string {
	return fmt.Sprintf("&{ipEndpoint:%+v hostHeader:%s}", gs.ipEndpoint, gs.hostHeader)
}
