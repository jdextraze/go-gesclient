package internal

import (
	"encoding/json"
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"time"
)

type ClusterDnsEndpointDiscoverer struct {
	clusterDns              string
	maxDiscoverAttemps      int
	managerExternalHttpPort int
	gossipSeeds             []*client.GossipSeed
	gossipTimeout           time.Duration
	oldGossip               *messages.ClusterInfoDto
}

func NewClusterDnsEndPointDiscoverer(
	clusterDns string,
	maxDiscoverAttemps int,
	managerExternalHttpPort int,
	gossipSeeds []*client.GossipSeed,
	gossipTimeout time.Duration,
) *ClusterDnsEndpointDiscoverer {
	return &ClusterDnsEndpointDiscoverer{
		clusterDns:              clusterDns,
		maxDiscoverAttemps:      maxDiscoverAttemps,
		managerExternalHttpPort: managerExternalHttpPort,
		gossipSeeds:             gossipSeeds,
		gossipTimeout:           gossipTimeout,
	}
}

func (d *ClusterDnsEndpointDiscoverer) DiscoverAsync(failedTcpEndpoint net.Addr) *tasks.Task {
	return tasks.NewStarted(func() (interface{}, error) {
		for attempt := 1; attempt <= d.maxDiscoverAttemps; attempt++ {
			endPoints, err := d.discoverEndpoint(failedTcpEndpoint)
			if err != nil {
				log.Infof("Discovering attempt %d/%d failed with error: %v.", attempt, d.maxDiscoverAttemps, err)
			} else if endPoints != nil {
				log.Infof("Discovering attempt %d/%d successful: best candidate is %s.", attempt, d.maxDiscoverAttemps,
					endPoints)
				return endPoints, nil
			} else {
				log.Infof("Discovering attempt %d/%d failed: no candidate found.", attempt, d.maxDiscoverAttemps)
			}
			time.Sleep(500 * time.Millisecond)
		}
		return nil, fmt.Errorf("Failed to discover candidate in %d attemps", d.maxDiscoverAttemps)
	})
}

func (d *ClusterDnsEndpointDiscoverer) discoverEndpoint(failedTcpEndpoint net.Addr) (*NodeEndpoints, error) {
	oldGossip := d.oldGossip
	var gossipCandidates []*client.GossipSeed
	if oldGossip != nil {
		gossipCandidates = d.getGossipCandidatesFromOldGossip(oldGossip, failedTcpEndpoint)
	} else {
		var err error
		gossipCandidates, err = d.getGossipCandidatesFromDns()
		if err != nil {
			return nil, err
		}
	}
	for _, gc := range gossipCandidates {
		gossip := d.tryGetGossipFrom(gc)
		if gossip == nil || gossip.Members == nil || len(gossip.Members) == 0 {
			continue
		}

		bestNode := d.tryDetermineBestNode(gossip.Members)
		if bestNode != nil {
			d.oldGossip = gossip
			return bestNode, nil
		}
	}
	return nil, nil
}

func (d *ClusterDnsEndpointDiscoverer) getGossipCandidatesFromOldGossip(
	oldGossip *messages.ClusterInfoDto,
	failedTcpEndpoint net.Addr,
) []*client.GossipSeed {
	if failedTcpEndpoint == nil {
		return d.arrangeGossipCandidates(oldGossip.Members)
	}
	var ip string
	var port int
	fmt.Sscanf(failedTcpEndpoint.String(), "%s:%d", &ip, &port)
	candidates := []*messages.MemberInfoDto{}
	for _, g := range oldGossip.Members {
		if !(g.ExternalTcpPort == port && g.ExternalTcpIp == ip) {
			candidates = append(candidates, g)
		}
	}
	return d.arrangeGossipCandidates(candidates)
}

func (d *ClusterDnsEndpointDiscoverer) getGossipCandidatesFromDns() ([]*client.GossipSeed, error) {
	var endpoints []*client.GossipSeed
	if d.gossipSeeds != nil && len(d.gossipSeeds) > 0 {
		endpoints = d.gossipSeeds
	} else {
		ipAddresses, err := d.resolveDns(d.clusterDns)
		if err != nil {
			return nil, err
		}
		endpoints = make([]*client.GossipSeed, len(ipAddresses))
		for i, ip := range ipAddresses {
			endpoints[i] = client.NewGossipSeed(&net.TCPAddr{IP: ip, Port: d.managerExternalHttpPort}, "")
		}
	}
	d.randomShuffle(endpoints, 0, len(endpoints)-1)
	return endpoints, nil
}

func (d *ClusterDnsEndpointDiscoverer) resolveDns(dns string) ([]net.IP, error) {
	ips, err := net.LookupIP(dns)
	if err != nil {
		return nil, fmt.Errorf("Error while resolving DNS entry '%s': %v", dns, err)
	}
	if ips == nil || len(ips) == 0 {
		return nil, fmt.Errorf("DNS entry '%s' resolved into empty list.", dns)
	}
	return ips, nil
}

func (d *ClusterDnsEndpointDiscoverer) tryGetGossipFrom(endpoint *client.GossipSeed) *messages.ClusterInfoDto {
	url := fmt.Sprintf("http://%s/gossip?format=json", endpoint.IpEndpoint().String())
	log.Infof("Trying to get gossip from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	clusterInfoDto := &messages.ClusterInfoDto{}
	json.Unmarshal(data, clusterInfoDto)
	return clusterInfoDto
}

func (d *ClusterDnsEndpointDiscoverer) tryDetermineBestNode(members []*messages.MemberInfoDto) *NodeEndpoints {
	var node *messages.MemberInfoDto
	sort.Sort(byStateDescending(members))
	for _, n := range members {
		if !n.IsAlive || n.State == messages.VNodeState_Manager || n.State == messages.VNodeState_ShuttingDown ||
			n.State == messages.VNodeState_Shutdown {
			continue
		}
		node = n
		break
	}
	normTcp := &net.TCPAddr{IP: net.ParseIP(node.ExternalTcpIp), Port: node.ExternalTcpPort}
	var secTcp net.Addr
	if node.ExternalSecureTcpPort > 0 {
		secTcp = &net.TCPAddr{IP: net.ParseIP(node.ExternalTcpIp), Port: node.ExternalSecureTcpPort}
	}
	log.Infof("Discovering: found best choice [%s,%s] (%s)", normTcp, secTcp, node.State)
	return NewNodeEndpoints(normTcp, secTcp)
}

func (d *ClusterDnsEndpointDiscoverer) arrangeGossipCandidates(
	members []*messages.MemberInfoDto,
) []*client.GossipSeed {
	result := make([]*client.GossipSeed, len(members))
	i := -1
	j := len(members)
	for k := 0; k < len(members); k++ {
		gossipSeed := client.NewGossipSeed(
			&net.TCPAddr{IP: net.ParseIP(members[k].ExternalHttpIp), Port: members[k].ExternalHttpPort}, "")
		if members[k].State == messages.VNodeState_Manager {
			j -= 1
			result[j] = gossipSeed
		} else {
			i += 1
			result[i] = gossipSeed
		}
	}
	d.randomShuffle(result, 0, i)
	d.randomShuffle(result, j, len(members)-1)
	return nil
}

func (d *ClusterDnsEndpointDiscoverer) randomShuffle(arr []*client.GossipSeed, i int, j int) {
	if i >= j {
		return
	}
	perm := rand.Perm(j + 1)
	for k := i; k <= j; k++ {
		index := perm[k]
		tmp := arr[index]
		arr[index] = arr[k]
		arr[k] = tmp
	}
}

type byStateDescending []*messages.MemberInfoDto

func (a byStateDescending) Len() int           { return len(a) }
func (a byStateDescending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStateDescending) Less(i, j int) bool { return a[i].State < a[j].State }
