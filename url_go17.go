// +build !go1.8

package gesclient

import (
	"net/url"
	"strconv"
	"strings"
)

func getUrlPort(u *url.URL) int {
	hostport := u.Host
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return 0
	}
	if i := strings.Index(hostport, "]:"); i != -1 {
		port, _ := strconv.Atoi(hostport[i+len("]:"):])
		return port
	}
	if strings.Contains(hostport, "]") {
		return 0
	}
	port, _ := strconv.Atoi(hostport[colon+len(":"):])
	return port
}

func getUrlHostname(u *url.URL) string {
	hostport := u.Host
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}
