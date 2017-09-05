package client

import (
	"github.com/satori/go.uuid"
)

type Operation interface {
	CreateNetworkPackage(correlationId uuid.UUID) (*Package, error)
	InspectPackage(p *Package) (*InspectionResult, error)
	Fail(err error) error
}
