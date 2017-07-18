package client

import "net"

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
