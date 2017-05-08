package gesclient

import (
	"net/url"
	"fmt"
	"net"
	"github.com/jdextraze/go-gesclient/internal"
	"github.com/jdextraze/go-gesclient/models"
)

func Create(settings *models.ConnectionSettings, uri *url.URL, name string) (models.Connection, error) {
	var scheme string
	var connectionSettings models.ConnectionSettings

	if uri == nil {
		scheme = ""
	} else {
		scheme = uri.Scheme
	}

	if settings == nil {
		connectionSettings = *models.DefaultConnectionSettings
	} else {
		connectionSettings = *settings
	}

	credentials := getCredentialsFromUri(uri)
	if credentials != nil {
		connectionSettings.DefaultUserCredentials = credentials
	}

	//if scheme == "discover" {
	//	// TODO handle errors
	//	port, _ := strconv.Atoi(uri.Port())
	//	clusterSettings, _ := newClusterSettings(uri.Host, connectionSettings.maxDiscoverAttempts, port,
	//		nil, connectionSettings.GossipTimeout())
	//
	//	endPointDiscoverer := newClusterDnsEndPointDiscoverer(connectionSettings.Log,
	//		clusterSettings.ClusterDns(),
	//		clusterSettings.MaxDiscoverAttempts(),
	//		clusterSettings.ExternalGossipPort(),
	//		clusterSettings.GossipSeeds(),
	//		clusterSettings.GossipTimeout())
	//}

	if scheme == "internal" {
		// TODO handle error
		tcpEndpoint, _ := net.ResolveTCPAddr("internal", uri.Host)
		discoverer, _ := internal.NewStaticEndpointDiscoverer(tcpEndpoint, connectionSettings.UseSslConnection())
		return internal.NewConnection(connectionSettings, nil, discoverer, name), nil
	} else {
		return nil, fmt.Errorf("Invalid scheme for connection '%s'", scheme)
	}
}

func getCredentialsFromUri(uri *url.URL) *models.UserCredentials {
	if uri == nil || uri.User == nil {
		return nil
	}
	password, _ := uri.User.Password()
	return models.NewUserCredentials(uri.User.Username(), password)
}
