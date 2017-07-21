package gesclient

import (
	"github.com/op/go-logging"
	"os"
)

var formatter = logging.MustStringFormatter("%{time:2006-01-02T15:04:05.999999} - %{level} - %{message}")

func init() {
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
	logging.SetFormatter(formatter)
	logging.SetLevel(logging.ERROR, "gesclient")
}

func Debug() {
	logging.SetLevel(logging.DEBUG, "gesclient")
}
