package gesclient

import (
	"github.com/satori/go.uuid"
)

const (
	tcpPacketContentLengthSize = 4
	tcpPacketCommandPos        = 0
	tcpPacketAuthFlagPos       = tcpPacketCommandPos + 1
	tcpPacketCorrelationIdPos  = tcpPacketAuthFlagPos + 1
	tcpPacketPayloadPos        = tcpPacketCorrelationIdPos + 16
)

const (
	tcpFlagsNone          byte = 0x00
	tcpFlagsAuthenticated byte = 0x01
)

type tcpPacket struct {
	Command       tcpCommand
	flags         byte
	CorrelationId uuid.UUID
	Payload       []byte
	username      string
	password      string
}

func newTcpPacket(
	cmd tcpCommand,
	flags byte,
	correlationId uuid.UUID,
	payload []byte,
	userCredentials *UserCredentials,
) *tcpPacket {
	var (
		username string
		password string
	)
	if userCredentials != nil {
		username = userCredentials.Username()
		password = userCredentials.Password()
	}
	return &tcpPacket{cmd, flags, correlationId, payload, username, password}
}

func tcpPacketFromBytes(data []byte) *tcpPacket {
	if len(data) < tcpPacketPayloadPos {
		return nil
	}
	command := tcpCommand(data[tcpPacketCommandPos])
	flags := data[tcpPacketAuthFlagPos]
	correlationId, _ := uuid.FromBytes(
		data[tcpPacketCorrelationIdPos:tcpPacketPayloadPos])
	payload := data[tcpPacketPayloadPos:]
	return &tcpPacket{command, flags, correlationId, payload, "", ""}
}

func (p *tcpPacket) Bytes() []byte {
	contentLength := p.Size()
	b := make([]byte, contentLength)
	b[tcpPacketCommandPos] = byte(p.Command)
	b[tcpPacketAuthFlagPos] = byte(p.flags)
	copy(b[tcpPacketCorrelationIdPos:], p.CorrelationId.Bytes())
	pos := tcpPacketPayloadPos
	if p.flags&tcpFlagsAuthenticated != 0 {
		b[pos] = byte(len(p.username))
		pos++
		copy(b[pos:], p.username)
		pos += len(p.username)
		b[pos] = byte(len(p.password))
		pos++
		copy(b[pos:], p.password)
		pos += len(p.password)
	}
	copy(b[pos:], p.Payload)
	return b
}

func (p *tcpPacket) Size() int32 {
	authLen := 0
	if p.flags&tcpFlagsAuthenticated != 0 {
		authLen = 2 + len(p.username) + len(p.password)
	}
	return int32(len(p.Payload) + tcpPacketPayloadPos + authLen)
}
