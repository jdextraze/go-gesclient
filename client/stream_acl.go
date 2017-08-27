package client

import (
	"fmt"
	"strings"
)

type streamAcl struct {
	readRoles      []string
	writeRoles     []string
	deleteRoles    []string
	metaReadRoles  []string
	metaWriteRoles []string
}

func NewStreamAcl(
	readRoles []string,
	writeRoles []string,
	deleteRoles []string,
	metaReadRoles []string,
	metaWriteRoles []string,
) *streamAcl {
	return &streamAcl{
		readRoles:      readRoles,
		writeRoles:     writeRoles,
		deleteRoles:    deleteRoles,
		metaReadRoles:  metaReadRoles,
		metaWriteRoles: metaWriteRoles,
	}
}

func (x *streamAcl) ReadRoles() []string { return x.readRoles }

func (x *streamAcl) WriteRoles() []string { return x.writeRoles }

func (x *streamAcl) DeleteRoles() []string { return x.deleteRoles }

func (x *streamAcl) MetaReadRoles() []string { return x.metaReadRoles }

func (x *streamAcl) MetaWriteRoles() []string { return x.metaWriteRoles }

func (x *streamAcl) String() string {
	return fmt.Sprintf("Read: %s, Write: %s, Delete: %s, MetaRead: %s, MetaWrite: %s",
		streamAcl_GetRolesAsStringOrNil(x.readRoles),
		streamAcl_GetRolesAsStringOrNil(x.writeRoles),
		streamAcl_GetRolesAsStringOrNil(x.deleteRoles),
		streamAcl_GetRolesAsStringOrNil(x.metaReadRoles),
		streamAcl_GetRolesAsStringOrNil(x.metaWriteRoles))
}

func streamAcl_GetRolesAsStringOrNil(roles []string) string {
	if roles == nil {
		return "<nil>"
	}
	return "[" + strings.Join(roles, ",") + "]"
}
