// +build go1.8

package gesclient

import (
	"net/url"
	"strconv"
)

func getUrlPort(u *url.URL) int {
	port, _ := strconv.Atoi(u.Port())
	return port
}

func getUrlHostname(u *url.URL) string {
	return u.Hostname()
}
