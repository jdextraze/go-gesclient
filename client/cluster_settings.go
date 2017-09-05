package client

import (
	"time"
)

type ClusterSettings struct {
	clusterDns          string
	maxDiscoverAttempts int
	externalGossipPort  int
	gossipSeeds         []*GossipSeed
	gossipTimeout       time.Duration
}

func NewClusterSettings(
	clusterDns string,
	maxDiscoverAttempts int,
	externalGossipPort int,
	gossipSeeds []*GossipSeed,
	gossipTimeout time.Duration,
) *ClusterSettings {
	if gossipSeeds == nil && clusterDns == "" {
		panic("clusterDns must be present")
	}
	if maxDiscoverAttempts < -1 {
		panic("maxDiscoverAttempts value is out of range. Allowed range: [-1, infinity].")
	}
	if gossipSeeds == nil && externalGossipPort <= 0 {
		panic("externalGossipPort must be positive")
	}

	return &ClusterSettings{
		clusterDns:          clusterDns,
		maxDiscoverAttempts: maxDiscoverAttempts,
		externalGossipPort:  externalGossipPort,
		gossipSeeds:         gossipSeeds,
		gossipTimeout:       gossipTimeout,
	}
}

func (cs *ClusterSettings) ClusterDns() string { return cs.clusterDns }

func (cs *ClusterSettings) MaxDiscoverAttempts() int { return cs.maxDiscoverAttempts }

func (cs *ClusterSettings) ExternalGossipPort() int { return cs.externalGossipPort }

func (cs *ClusterSettings) GossipSeeds() []*GossipSeed { return cs.gossipSeeds }

func (cs *ClusterSettings) GossipTimeout() time.Duration { return cs.gossipTimeout }
