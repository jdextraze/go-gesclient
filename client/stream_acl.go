package client

import (
	"encoding/json"
	"errors"
)

type StreamAcl struct {
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
) *StreamAcl {
	return &StreamAcl{
		readRoles:      readRoles,
		writeRoles:     writeRoles,
		deleteRoles:    deleteRoles,
		metaReadRoles:  metaReadRoles,
		metaWriteRoles: metaWriteRoles,
	}
}

func (x *StreamAcl) ReadRoles() []string { return x.readRoles }

func (x *StreamAcl) WriteRoles() []string { return x.writeRoles }

func (x *StreamAcl) DeleteRoles() []string { return x.deleteRoles }

func (x *StreamAcl) MetaReadRoles() []string { return x.metaReadRoles }

func (x *StreamAcl) MetaWriteRoles() []string { return x.metaWriteRoles }

type stringOrStringArray []string

func (s stringOrStringArray) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

func (s *stringOrStringArray) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]string)(s))
	} else if data[0] == '"' {
		*s = stringOrStringArray{""}
		return json.Unmarshal(data, &((*s)[0]))
	}
	return errors.New("invalid format")
}

type streamAclJson struct {
	ReadRoles      stringOrStringArray `json:"$r,omitempty"`
	WriteRoles     stringOrStringArray `json:"$w,omitempty"`
	DeleteRoles    stringOrStringArray `json:"$d,omitempty"`
	MetaReadRoles  stringOrStringArray `json:"$mr,omitempty"`
	MetaWriteRoles stringOrStringArray `json:"$mw,omitempty"`
}

func (x *StreamAcl) MarshalJSON() ([]byte, error) {
	return json.Marshal(streamAclJson{
		ReadRoles:      x.readRoles,
		WriteRoles:     x.writeRoles,
		DeleteRoles:    x.deleteRoles,
		MetaReadRoles:  x.metaReadRoles,
		MetaWriteRoles: x.metaWriteRoles,
	})
}

func (x *StreamAcl) UnmarshalJSON(data []byte) error {
	o := streamAclJson{}
	if err := json.Unmarshal(data, &o); err != nil {
		return err
	}
	x.readRoles = o.ReadRoles
	x.writeRoles = o.WriteRoles
	x.deleteRoles = o.DeleteRoles
	x.metaReadRoles = o.MetaReadRoles
	x.metaWriteRoles = o.MetaWriteRoles
	return nil
}
