package internal

import (
	"github.com/jdextraze/go-gesclient/models"
	"time"
	"net"
)

type ClusterDnsEndpointDiscoverer struct {
	clusterDns              string
	maxDiscoverAttemps      int
	managerExternalHttpPort int
	gossipSeeds             []*models.GossipSeed
	gossipTimeout           time.Duration
}

func NewClusterDnsEndPointDiscoverer(
	clusterDns string,
	maxDiscoverAttemps int,
	managerExternalHttpPort int,
	gossipSeeds []*models.GossipSeed,
	gossipTimeout time.Duration,
) *ClusterDnsEndpointDiscoverer {
	panic("TODO") // TODO
	return &ClusterDnsEndpointDiscoverer{
		clusterDns: clusterDns,
		maxDiscoverAttemps: maxDiscoverAttemps,
		managerExternalHttpPort: managerExternalHttpPort,
		gossipSeeds: gossipSeeds,
		gossipTimeout: gossipTimeout,
	}
}

func (d *ClusterDnsEndpointDiscoverer) DiscoverAsync(ipEndpoint net.Addr) chan *EndpointDiscovererResult {
	return nil
}
