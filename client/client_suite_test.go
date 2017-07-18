package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/op/go-logging"
	"testing"
)

func TestGoGesclient(t *testing.T) {
	logging.SetLevel(logging.CRITICAL, "gesclient")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}
