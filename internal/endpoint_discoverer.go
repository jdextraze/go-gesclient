package internal

import (
	"net"
)

type EndpointDiscovererResult struct {
	nodeEndpoints *NodeEndpoints
	error         error
}

type EndpointDiscoverer interface {
	DiscoverAsync(ipEndpoint net.Addr) chan *EndpointDiscovererResult
}

type staticEndpointDiscoverer struct {
	result chan *EndpointDiscovererResult
}

func NewStaticEndpointDiscoverer(ipEndpoint net.Addr, isSsl bool) *staticEndpointDiscoverer {
	if ipEndpoint == nil {
		panic("ipEndpoint is nil")
	}
	ch := make(chan *EndpointDiscovererResult, 1)
	var nodeEndpoints *NodeEndpoints
	if isSsl {
		nodeEndpoints = NewNodeEndpoints(nil, ipEndpoint)
	} else {
		nodeEndpoints = NewNodeEndpoints(ipEndpoint, nil)
	}
	go func() {
		for {
			ch <- &EndpointDiscovererResult{
				nodeEndpoints: nodeEndpoints,
			}
		}
	}()
	return &staticEndpointDiscoverer{
		result: ch,
	}
}

func (d *staticEndpointDiscoverer) DiscoverAsync(ipEndpoint net.Addr) chan *EndpointDiscovererResult {
	return d.result
}
