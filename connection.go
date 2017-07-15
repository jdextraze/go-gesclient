package gesclient

import (
	"net/url"
	"fmt"
	"net"
	"github.com/jdextraze/go-gesclient/internal"
	"github.com/jdextraze/go-gesclient/models"
	"strconv"
)

func Create(settings *models.ConnectionSettings, uri *url.URL, name string) (models.Connection, error) {
	var scheme string
	var connectionSettings *models.ConnectionSettings

	if uri == nil {
		scheme = ""
	} else {
		scheme = uri.Scheme
	}

	if settings == nil {
		connectionSettings = models.DefaultConnectionSettings
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
		clusterSettings := models.NewClusterSettings(uri.Host, connectionSettings.MaxDiscoverAttempts(), port,
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
		clusterSettings := models.NewClusterSettings("", connectionSettings.MaxDiscoverAttempts(), 0,
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

func getCredentialsFromUri(uri *url.URL) *models.UserCredentials {
	if uri == nil || uri.User == nil {
		return nil
	}
	password, _ := uri.User.Password()
	return models.NewUserCredentials(uri.User.Username(), password)
}
