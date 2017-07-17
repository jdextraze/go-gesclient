package client

import (
	"net"
	"time"
)

type ConnectionSettingsBuilder struct {
	verboseLogging              bool
	maxQueueSize                int
	maxConcurrentItem           int
	maxRetries                  int
	maxReconnections            int
	requireMaster               bool
	reconnectionDelay           time.Duration
	operationTimeout            time.Duration
	operationTimeoutCheckPeriod time.Duration
	defaultUserCredentials      *UserCredentials
	useSslConnection            bool
	targetHost                  string
	validateServer              bool
	failOnNoServerResponse      bool
	heartbeatInterval           time.Duration
	heartbeatTimeout            time.Duration
	clusterDns                  string
	maxDiscoverAttempts         int
	externalGossipPort          int
	gossipSeeds                 []*GossipSeed
	gossipTimeout               time.Duration
	clientConnectionTimeout     time.Duration
}

func CreateConnectionSettings() *ConnectionSettingsBuilder {
	return &ConnectionSettingsBuilder{
		verboseLogging:              false,
		maxQueueSize:                DefaultMaxQueueSize,
		maxConcurrentItem:           DefaultMaxConcurrentItems,
		maxRetries:                  DefaultMaxOperationRetries,
		maxReconnections:            DefaultMaxReconnections,
		requireMaster:               DefaultRequireMaster,
		reconnectionDelay:           DefaultReconnectionDelay,
		operationTimeout:            DefaultOperationTimeout,
		operationTimeoutCheckPeriod: DefaultOperationTimeoutCheckPeriod,
		defaultUserCredentials:      nil,
		useSslConnection:            false,
		targetHost:                  "",
		validateServer:              false,
		failOnNoServerResponse:      false,
		heartbeatInterval:           750 * time.Millisecond,
		heartbeatTimeout:            1500 * time.Millisecond,
		clusterDns:                  "",
		maxDiscoverAttempts:         DefaultMaxClusterDiscoverAttempts,
		externalGossipPort:          DefaultClusterManagerExternalHttpPort,
		gossipSeeds:                 nil,
		gossipTimeout:               1 * time.Second,
		clientConnectionTimeout:     5 * time.Second,
	}
}

func (csb *ConnectionSettingsBuilder) EnableVerboseLogging() *ConnectionSettingsBuilder {
	csb.verboseLogging = true
	return csb
}

func (csb *ConnectionSettingsBuilder) LimitOperationsQueueTo(limit int) *ConnectionSettingsBuilder {
	csb.maxQueueSize = limit
	return csb
}

func (csb *ConnectionSettingsBuilder) LimitConcurrentOperationsTo(limit int) *ConnectionSettingsBuilder {
	csb.maxConcurrentItem = limit
	return csb
}

func (csb *ConnectionSettingsBuilder) LimitAttemptsForOperationTo(limit int) *ConnectionSettingsBuilder {
	csb.maxRetries = limit - 1
	return csb
}

func (csb *ConnectionSettingsBuilder) LimitRetriesForOperationTo(limit int) *ConnectionSettingsBuilder {
	csb.maxRetries = limit
	return csb
}

func (csb *ConnectionSettingsBuilder) KeepRetrying() *ConnectionSettingsBuilder {
	csb.maxRetries = -1
	return csb
}

func (csb *ConnectionSettingsBuilder) LimitReconnectionsTo(limit int) *ConnectionSettingsBuilder {
	csb.maxReconnections = limit
	return csb
}

func (csb *ConnectionSettingsBuilder) KeepReconnecting() *ConnectionSettingsBuilder {
	csb.maxReconnections = -1
	return csb
}

func (csb *ConnectionSettingsBuilder) PerformOnMasterOnly() *ConnectionSettingsBuilder {
	csb.requireMaster = true
	return csb
}

func (csb *ConnectionSettingsBuilder) PerformOnAnyNode() *ConnectionSettingsBuilder {
	csb.requireMaster = false
	return csb
}

func (csb *ConnectionSettingsBuilder) SetReconnectionDelayTo(delay time.Duration) *ConnectionSettingsBuilder {
	csb.reconnectionDelay = delay
	return csb
}

func (csb *ConnectionSettingsBuilder) SetOperationTimeoutTo(delay time.Duration) *ConnectionSettingsBuilder {
	csb.operationTimeout = delay
	return csb
}

func (csb *ConnectionSettingsBuilder) SetTimeoutCheckPeriodTo(period time.Duration) *ConnectionSettingsBuilder {
	csb.operationTimeoutCheckPeriod = period
	return csb
}

func (csb *ConnectionSettingsBuilder) SetDefaultUserCredentials(userCredentials *UserCredentials) *ConnectionSettingsBuilder {
	csb.defaultUserCredentials = userCredentials
	return csb
}

func (csb *ConnectionSettingsBuilder) UseSslConnection(targetHost string, validateServer bool) *ConnectionSettingsBuilder {
	csb.useSslConnection = true
	csb.targetHost = targetHost
	csb.validateServer = validateServer
	return csb
}

func (csb *ConnectionSettingsBuilder) FailOnNoServerResponse() *ConnectionSettingsBuilder {
	csb.failOnNoServerResponse = true
	return csb
}

func (csb *ConnectionSettingsBuilder) SetHeartbeatInterval(interval time.Duration) *ConnectionSettingsBuilder {
	csb.heartbeatInterval = interval
	return csb
}

func (csb *ConnectionSettingsBuilder) SetHeartbeatTimeout(timeout time.Duration) *ConnectionSettingsBuilder {
	csb.heartbeatTimeout = timeout
	return csb
}

func (csb *ConnectionSettingsBuilder) WithConnectionTimeoutOf(timeout time.Duration) *ConnectionSettingsBuilder {
	csb.clientConnectionTimeout = timeout
	return csb
}

func (csb *ConnectionSettingsBuilder) SetClusterDns(clusterDns string) *ConnectionSettingsBuilder {
	csb.clusterDns = clusterDns
	return csb
}

func (csb *ConnectionSettingsBuilder) SetMaxDiscoverAttempts(maxAttempts int) *ConnectionSettingsBuilder {
	csb.maxDiscoverAttempts = maxAttempts
	return csb
}

func (csb *ConnectionSettingsBuilder) SetGossipTimeout(timeout time.Duration) *ConnectionSettingsBuilder {
	csb.gossipTimeout = timeout
	return csb
}

func (csb *ConnectionSettingsBuilder) SetClusterGossipPort(port int) *ConnectionSettingsBuilder {
	csb.externalGossipPort = port
	return csb
}

func (csb *ConnectionSettingsBuilder) SetGossipSeedEndPoints(endpoints []*net.TCPAddr) *ConnectionSettingsBuilder {
	gossipSeeds := make([]*GossipSeed, len(endpoints))
	for i, endpoint := range endpoints {
		gossipSeeds[i] = NewGossipSeed(endpoint, "")
	}
	csb.gossipSeeds = gossipSeeds
	return csb
}

func (csb *ConnectionSettingsBuilder) SetGossipSeeds(gossipSeeds []*GossipSeed) *ConnectionSettingsBuilder {
	csb.gossipSeeds = gossipSeeds
	return csb
}

func (csb *ConnectionSettingsBuilder) Build() *ConnectionSettings {
	return newConnectionSettings(
		csb.verboseLogging,
		csb.maxQueueSize,
		csb.maxConcurrentItem,
		csb.maxRetries,
		csb.maxReconnections,
		csb.requireMaster,
		csb.reconnectionDelay,
		csb.operationTimeout,
		csb.operationTimeoutCheckPeriod,
		csb.defaultUserCredentials,
		csb.useSslConnection,
		csb.targetHost,
		csb.validateServer,
		csb.failOnNoServerResponse,
		csb.heartbeatInterval,
		csb.heartbeatTimeout,
		csb.clusterDns,
		csb.maxDiscoverAttempts,
		csb.externalGossipPort,
		csb.gossipSeeds,
		csb.gossipTimeout,
		csb.clientConnectionTimeout,
	)
}
