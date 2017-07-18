package gesclient

import "github.com/op/go-logging"

func init() {
	logging.SetLevel(logging.ERROR, "gesclient")
	logging.SetLevel(logging.ERROR, "internal")
	logging.SetLevel(logging.ERROR, "operations")
	logging.SetLevel(logging.ERROR, "client")
}

func Debug() {
	logging.SetLevel(logging.DEBUG, "gesclient")
	logging.SetLevel(logging.DEBUG, "internal")
	logging.SetLevel(logging.DEBUG, "operations")
	logging.SetLevel(logging.DEBUG, "client")
}
