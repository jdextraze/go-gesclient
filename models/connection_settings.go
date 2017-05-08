package models

import (
	"github.com/op/go-logging"
	"time"
	"errors"
)

var DefaultConnectionSettings, _ = CreateConnectionSettings().Build()

type ConnectionSettings struct {
	log                         logging.Logger
	verboseLogging              bool
	maxQueueSize                int
	maxConcurrentItem           int
	maxRetries                  int
	maxReconnections            int
	requireMaster               bool
	reconnectionDelay           time.Duration
	operationTimeout            time.Duration
	operationTimeoutCheckPeriod time.Duration
	DefaultUserCredentials      *UserCredentials
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

func newConnectionSettings(
	log *logging.Logger,
	verboseLogging bool,
	maxQueueSize int,
	maxConcurrentItem int,
	maxRetries int,
	maxReconnections int,
	requireMaster bool,
	reconnectionDelay time.Duration,
	operationTimeout time.Duration,
	operationTimeoutCheckPeriod time.Duration,
	defaultUserCredentials *UserCredentials,
	useSslConnection bool,
	targetHost string,
	validateService bool,
	failOnNoServerResponse bool,
	heartbeatInterval time.Duration,
	heartbeatTimeout time.Duration,
	clusterDns string,
	maxDiscoverAttempts int,
	externalGossipPort int,
	gossipSeeds []*GossipSeed,
	gossipTimeout time.Duration,
	clientConnectionTimeout time.Duration,
) (*ConnectionSettings, error) {
	if log == nil {
		return nil, errors.New("Logger should not be nil")
	}
	if maxQueueSize <= 0 {
		return nil, errors.New("maxQueueSize should be positive")
	}
	if maxConcurrentItem <= 0 {
		return nil, errors.New("maxConcurrentItem should be positive")
	}
	if maxRetries < -1 {
		return nil, errors.New("maxRetries is out of range. Allowed range: [-1, infinity]")
	}
	if maxReconnections < -1 {
		return nil, errors.New("maxReconnections is out of range. Allowed range: [-1, infinity]")
	}
	if useSslConnection && targetHost == "" {
		return nil, errors.New("targetHost must be present")
	}
	return &ConnectionSettings{
		log:                         *log,
		verboseLogging:              verboseLogging,
		maxQueueSize:                maxQueueSize,
		maxConcurrentItem:           maxConcurrentItem,
		maxRetries:                  maxRetries,
		maxReconnections:            maxReconnections,
		requireMaster:               requireMaster,
		reconnectionDelay:           reconnectionDelay,
		operationTimeout:            operationTimeout,
		operationTimeoutCheckPeriod: operationTimeoutCheckPeriod,
		DefaultUserCredentials:      defaultUserCredentials,
		useSslConnection:            useSslConnection,
		targetHost:                  targetHost,
		validateServer:              validateService,
		failOnNoServerResponse:      failOnNoServerResponse,
		heartbeatInterval:           heartbeatInterval,
		heartbeatTimeout:            heartbeatTimeout,
		clusterDns:                  clusterDns,
		maxDiscoverAttempts:         maxDiscoverAttempts,
		externalGossipPort:          externalGossipPort,
		gossipSeeds:                 gossipSeeds,
		gossipTimeout:               gossipTimeout,
		clientConnectionTimeout:     clientConnectionTimeout,
	}, nil
}

func (cs *ConnectionSettings) Log() logging.Logger {
	return cs.log
}

func (cs *ConnectionSettings) VerboseLogging() bool {
	return cs.verboseLogging
}

func (cs *ConnectionSettings) MaxQueueSize() int {
	return cs.maxQueueSize
}

func (cs *ConnectionSettings) MaxConcurrentItem() int {
	return cs.maxConcurrentItem
}

func (cs *ConnectionSettings) MaxRetries() int {
	return cs.maxRetries
}

func (cs *ConnectionSettings) MaxReconnections() int {
	return cs.maxReconnections
}

func (cs *ConnectionSettings) RequireMaster() bool {
	return cs.requireMaster
}

func (cs *ConnectionSettings) ReconnectionDelay() time.Duration {
	return cs.reconnectionDelay
}

func (cs *ConnectionSettings) OperationTimeout() time.Duration {
	return cs.operationTimeout
}

func (cs *ConnectionSettings) OperationTimeoutCheckPeriod() time.Duration {
	return cs.operationTimeoutCheckPeriod
}

func (cs *ConnectionSettings) UseSslConnection() bool {
	return cs.useSslConnection
}

func (cs *ConnectionSettings) TargetHost() string {
	return cs.targetHost
}

func (cs *ConnectionSettings) ValidateService() bool {
	return cs.validateServer
}

func (cs *ConnectionSettings) FailOnNoServerResponse() bool {
	return cs.failOnNoServerResponse
}

func (cs *ConnectionSettings) HeartbeatInterval() time.Duration {
	return cs.heartbeatInterval
}

func (cs *ConnectionSettings) HeartbeatTimeout() time.Duration {
	return cs.heartbeatInterval
}

func (cs *ConnectionSettings) ClusterDns() string {
	return cs.clusterDns
}

func (cs *ConnectionSettings) MaxDiscoverAttempts() int {
	return cs.maxDiscoverAttempts
}

func (cs *ConnectionSettings) ExternalGossipPort() int {
	return cs.externalGossipPort
}

func (cs *ConnectionSettings) GossipSeeds() []*GossipSeed {
	return cs.gossipSeeds
}

func (cs *ConnectionSettings) GossipTimeout() time.Duration {
	return cs.gossipTimeout
}

func (cs *ConnectionSettings) ClientConnectionTimeout() time.Duration {
	return cs.clientConnectionTimeout
}
