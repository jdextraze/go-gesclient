package client

import (
	"fmt"
	"github.com/satori/go.uuid"
)

type Operation interface {
	fmt.Stringer
	CreateNetworkPackage(correlationId uuid.UUID) (*Package, error)
	InspectPackage(p *Package) *InspectionResult
	Fail(err error)
}
