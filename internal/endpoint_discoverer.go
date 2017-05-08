package internal

import (
	"net"
	"errors"
)

type EndpointDiscovererResult struct {
	nodeEndpoints *NodeEndpoints
	error         error
}

type EndpointDiscoverer interface {
	DiscoverAsync(ipEndpoint *net.TCPAddr) chan *EndpointDiscovererResult
}

type staticEndpointDiscoverer struct {
	result chan *EndpointDiscovererResult
}

func NewStaticEndpointDiscoverer(ipEndpoint *net.TCPAddr, isSsl bool) (discoverer *staticEndpointDiscoverer, err error) {
	if ipEndpoint == nil {
		err = errors.New("ipEndpoint is nil")
		return
	}
	ch := make(chan *EndpointDiscovererResult, 1)
	var nodeEndpoints *NodeEndpoints
	if isSsl {
		nodeEndpoints, err = NewNodeEndpoints(nil, ipEndpoint)
	} else {
		nodeEndpoints, err = NewNodeEndpoints(ipEndpoint, nil)
	}
	if err != nil {
		return
	}
	ch <- &EndpointDiscovererResult{
		nodeEndpoints: nodeEndpoints,
	}
	discoverer = &staticEndpointDiscoverer{
		result: ch,
	}
	return
}

func (d *staticEndpointDiscoverer) DiscoverAsync(ipEndpoint *net.TCPAddr) chan *EndpointDiscovererResult {
	return d.result
}
