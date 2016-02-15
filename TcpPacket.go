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

type tcpPacket struct {
    Command       tcpCommand
    AuthFlag      byte
    CorrelationId uuid.UUID
    Payload       []byte
}

func newTcpPacket(
    cmd tcpCommand,
    authFlag byte,
    correlationId uuid.UUID,
    payload []byte,
) *tcpPacket {
    return &tcpPacket{cmd, authFlag, correlationId, payload}
}

func tcpPacketFromBytes(data []byte) *tcpPacket {
    if len(data) < tcpPacketPayloadPos {
        return nil
    }
    command := tcpCommand(data[tcpPacketCommandPos])
    authFlag := data[tcpPacketAuthFlagPos]
    correlationId, _ := uuid.FromBytes(
        data[tcpPacketCorrelationIdPos:tcpPacketPayloadPos])
    payload := data[tcpPacketPayloadPos:]
    return &tcpPacket{command, authFlag, correlationId, payload}
}

func (p *tcpPacket) Bytes() []byte {
    contentLength := p.Size()
    b := make([]byte, contentLength)
    b[tcpPacketCommandPos] = byte(p.Command)
    b[tcpPacketAuthFlagPos] = byte(p.AuthFlag)
    copy(b[tcpPacketCorrelationIdPos:], p.CorrelationId.Bytes())
    if contentLength > tcpPacketPayloadPos {
        copy(b[tcpPacketPayloadPos:], p.Payload)
    }
    return b
}

func (p *tcpPacket) Size() int32 {
    return int32(len(p.Payload) + tcpPacketPayloadPos)
}
