package gesclient

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/internal"
	"net"
	"net/url"
	"strconv"
)

func Create(settings *client.ConnectionSettings, uri *url.URL, name string) (client.Connection, error) {
	var scheme string
	var connectionSettings *client.ConnectionSettings

	if uri == nil {
		scheme = ""
	} else {
		scheme = uri.Scheme
	}

	if settings == nil {
		connectionSettings = client.DefaultConnectionSettings
	} else {
		connectionSettings = settings
	}

	credentials := getCredentialsFromUri(uri)
	if credentials != nil {
		connectionSettings.DefaultUserCredentials = credentials
	}

	var endPointDiscoverer internal.EndpointDiscoverer
	if scheme == "discover" {
		port, _ := strconv.Atoi(uri.Port())
		clusterSettings := client.NewClusterSettings(uri.Hostname(), connectionSettings.MaxDiscoverAttempts(), port,
			nil, connectionSettings.GossipTimeout())

		endPointDiscoverer = internal.NewClusterDnsEndPointDiscoverer(
			clusterSettings.ClusterDns(),
			clusterSettings.MaxDiscoverAttempts(),
			clusterSettings.ExternalGossipPort(),
			clusterSettings.GossipSeeds(),
			clusterSettings.GossipTimeout())
	} else if scheme == "tcp" {
		tcpEndpoint, err := net.ResolveTCPAddr("tcp", uri.Host)
		if err != nil {
			return nil, err
		}
		endPointDiscoverer = internal.NewStaticEndpointDiscoverer(tcpEndpoint, connectionSettings.UseSslConnection())
	} else if connectionSettings.GossipSeeds() != nil && len(connectionSettings.GossipSeeds()) > 0 {
		clusterSettings := client.NewClusterSettings("", connectionSettings.MaxDiscoverAttempts(), 0,
			connectionSettings.GossipSeeds(), connectionSettings.GossipTimeout())

		endPointDiscoverer = internal.NewClusterDnsEndPointDiscoverer(
			clusterSettings.ClusterDns(),
			clusterSettings.MaxDiscoverAttempts(),
			clusterSettings.ExternalGossipPort(),
			clusterSettings.GossipSeeds(),
			clusterSettings.GossipTimeout())
	} else {
		return nil, fmt.Errorf("Invalid scheme for connection '%s'", scheme)
	}
	return internal.NewConnection(connectionSettings, nil, endPointDiscoverer, name), nil
}

func getCredentialsFromUri(uri *url.URL) *client.UserCredentials {
	if uri == nil || uri.User == nil {
		return nil
	}
	password, _ := uri.User.Password()
	return client.NewUserCredentials(uri.User.Username(), password)
}
