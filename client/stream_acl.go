package client

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
