package gesclient

import "github.com/op/go-logging"

func init() {
	logging.SetLevel(logging.ERROR, "gesclient")
}

func Debug() {
	logging.SetLevel(logging.DEBUG, "gesclient")
}
