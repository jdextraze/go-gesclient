package client

import (
	"github.com/jdextraze/go-gesclient/pkg/uuid"
)

type Operation interface {
	CreateNetworkPackage(correlationId uuid.UUID) (*Package, error)
	InspectPackage(p *Package) (*InspectionResult, error)
	Fail(err error) error
}
