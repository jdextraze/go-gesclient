package models

import (
	"time"
	"errors"
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
) (*ClusterSettings, error) {
	if clusterDns == "" {
		return nil, errors.New("clusterDns must be present")
	}
	if maxDiscoverAttempts < -1 {
		return nil, errors.New("maxDiscoverAttempts value is out of range. Allowed range: [-1, infinity].")
	}
	if externalGossipPort <= 0 {
		return nil, errors.New("externalGossipPort must be positive")
	}

	return &ClusterSettings{
		clusterDns:          clusterDns,
		maxDiscoverAttempts: maxDiscoverAttempts,
		externalGossipPort:  externalGossipPort,
		gossipSeeds:         gossipSeeds,
		gossipTimeout:       gossipTimeout,
	}, nil
}

func (cs *ClusterSettings) ClusterDns() string { return cs.clusterDns }

func (cs *ClusterSettings) MaxDiscoverAttempts() int { return cs.maxDiscoverAttempts }

func (cs *ClusterSettings) ExternalGossipPort() int { return cs.externalGossipPort }

func (cs *ClusterSettings) GossipSeeds() []*GossipSeed { return cs.gossipSeeds }

func (cs *ClusterSettings) GossipTimeout() time.Duration { return cs.gossipTimeout }
