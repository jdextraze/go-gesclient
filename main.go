package gesclient

import "github.com/op/go-logging"

var log = logging.MustGetLogger("gesclient")

func init() {
	logging.SetLevel(logging.ERROR, "gesclient")
}

func Debug() {
	logging.SetLevel(logging.DEBUG, "gesclient")
}
