package client

import "github.com/gofrs/uuid"

type Operation interface {
	CreateNetworkPackage(correlationId uuid.UUID) (*Package, error)
	InspectPackage(p *Package) (*InspectionResult, error)
	Fail(err error) error
}
