package gesclient

import (
	. "bitbucket.org/jdextraze/go-gesclient/protobuf"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"github.com/satori/go.uuid"
	"io"
	"net"
	"reflect"
	"time"
    "encoding/binary"
    "bytes"
)

var log = logging.MustGetLogger("gesclient")

type Connection interface {
	Close() error
	CreateEvent(stream string, eventType string, isJson bool, data []byte,
		metadata []byte, expectedVersion int32) (*WriteEventsCompleted, error)
	CreateEventAsync(stream string, eventType string, isJson bool, data []byte,
		metadata []byte, expectedVersion int32) (chan *WriteEventsCompleted, error)
	ReadStreamEventsForward(stream string, start int32, max int32) (*ReadStreamEventsCompleted, error)
	ReadStreamEventsForwardAsync(stream string, start int32, max int32) (
        chan *ReadStreamEventsCompleted, error)
	SubscribeToStream(stream string) (*SubscriptionResult, error)
	SubscribeToStreamAsync(stream string) (chan *SubscriptionResult, error)
}

type connection struct {
	address      string
	conn         net.Conn
	asyncResults map[uuid.UUID]interface{}
	leftOver     []byte
	reconnect    bool
	connected    bool
}

func NewConnection(addr string) (Connection, error) {
	c := &connection{
		address:      addr,
		asyncResults: make(map[uuid.UUID]interface{}),
	}
	return c, c.connect()
}

func (c *connection) connect() error {
	log.Debug("Connecting to %s", c.address)
	conn, err := net.DialTimeout("tcp4", c.address, time.Second*3)
	if err != nil {
		return err
	}
	c.reconnect = true
	c.connected = true
	go func() {
		for {
			buff := make([]byte, 4096)
			n, err := conn.Read(buff)
			if err == nil {
				c.onData(buff[:n])
			} else if err == io.EOF {
				log.Debug("Disconnected from %s", c.address)
				break
			} else {
				log.Error("conn.Read:", err)
			}
		}

		c.conn.Close()
		c.connected = false

		for c.reconnect {
			err := c.connect()
			if err == nil {
				break
			}
			log.Error("Reconnect failed: %v", err)
			time.Sleep(time.Second)
		}
	}()
	c.conn = conn
	return nil
}

func (c *connection) assertConnected() error {
	if !c.connected {
		return errors.New("Not connected")
	}
	return nil
}

func (c *connection) Close() error {
	c.reconnect = false
	return c.conn.Close()
}

func (c *connection) CreateEvent(
	stream string,
	eventType string,
	isJson bool,
	data []byte,
	metadata []byte,
	expectedVersion int32,
) (*WriteEventsCompleted, error) {
	ch, err := c.CreateEventAsync(stream, eventType, isJson, data, metadata, expectedVersion)
	if err != nil {
		return nil, err
	}
	return <-ch, err
}

func (c *connection) CreateEventAsync(
	stream string,
	eventType string,
	isJson bool,
	data []byte,
	metadata []byte,
	expectedVersion int32,
) (chan *WriteEventsCompleted, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}

	var contentType int32
	if isJson {
		contentType = 1
	} else {
		contentType = 0
	}
	var metadataContentType int32 = 0
	requireMaster := false

	res := make(chan *WriteEventsCompleted)
	return res, c.sendCommand(
		tcpCommand_WriteEvents,
		&WriteEvents{
			EventStreamId:   &stream,
			ExpectedVersion: &expectedVersion,
			Events: []*NewEvent{
				&NewEvent{
					EventId:             uuid.NewV4().Bytes(),
					EventType:           &eventType,
					DataContentType:     &contentType,
					MetadataContentType: &metadataContentType,
					Data:                data,
					Metadata:            metadata,
				},
			},
			RequireMaster: &requireMaster,
		},
		res,
		uuid.Nil,
	)
}

func (c *connection) ReadStreamEventsForward(
	stream string,
	start int32,
	max int32,
) (*ReadStreamEventsCompleted, error) {
	ch, err := c.ReadStreamEventsForwardAsync(stream, start, max)
	if err != nil {
		return nil, err
	}
	return <-ch, nil
}

func (c *connection) ReadStreamEventsForwardAsync(
	stream string,
	start int32,
	max int32,
) (chan *ReadStreamEventsCompleted, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}

	no := false
	res := make(chan *ReadStreamEventsCompleted)
	return res, c.sendCommand(
		tcpCommand_ReadStreamEventsForward,
		&ReadStreamEvents{
			EventStreamId:   &stream,
			FromEventNumber: &start,
			MaxCount:        &max,
			ResolveLinkTos:  &no,
			RequireMaster:   &no,
		},
		res,
		uuid.Nil,
	)
}

func (c *connection) SubscribeToStream(stream string) (*SubscriptionResult, error) {
	ch, err := c.SubscribeToStreamAsync(stream)
	if err != nil {
		return nil, err
	}
	return <-ch, nil
}

func (c *connection) SubscribeToStreamAsync(stream string) (chan *SubscriptionResult, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}

	no := false
	res := make(chan *SubscriptionResult)
	return res, c.sendCommand(
		tcpCommand_SubscribeToStream,
		&SubscribeToStream{
			EventStreamId:  &stream,
			ResolveLinkTos: &no,
		},
		res,
		uuid.Nil,
	)
}

func (c *connection) sendCommand(
	tc tcpCommand,
	pb proto.Message,
	res interface{},
	correlationId uuid.UUID,
) error {
	payload, err := proto.Marshal(pb)
	if err != nil {
		closeIfChan(res)
		return err
	}

	if correlationId == uuid.Nil {
		correlationId = uuid.NewV4()
	}
	if err := c.sendTcpPacket(newTcpPacket(tc, 0, correlationId, payload)); err != nil {
		closeIfChan(res)
		return err
	}

	if res != nil {
		c.asyncResults[correlationId] = res
	}
	return nil
}

func (c *connection) onData(data []byte) {
	if c.leftOver != nil && len(c.leftOver) > 0 {
		data = append(c.leftOver, data...)
		c.leftOver = nil
	}

    dataLength := int32(len(data))
    if dataLength < tcpPacketContentLengthSize {
        c.leftOver = data
        return
    }
    var contentLength int32
    var buf = bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &contentLength)

    packetSize := contentLength + tcpPacketContentLengthSize
    if dataLength == packetSize {
        go c.process(tcpPacketFromBytes(data[tcpPacketContentLengthSize:]))
    } else if dataLength > packetSize {
        c.onData(data[:packetSize])
        c.onData(data[packetSize:])
    } else {
        c.leftOver = data
    }
}

func (c *connection) process(p *tcpPacket) {
	log.Info("Received Command: %s | CorrelationId: %s", p.Command, p.CorrelationId)

	asyncResult := c.asyncResults[p.CorrelationId]
	v := reflect.ValueOf(asyncResult)
	res := parsePayload(p.Command, p.Payload)
	resValue := reflect.ValueOf(res)
	if v.Kind() == reflect.Chan && resValue.Type() == v.Type().Elem() {
		v.Send(resValue)
		v.Close()
		delete(c.asyncResults, p.CorrelationId)
		return
	}

	switch p.Command {
	case tcpCommand_HeartbeatRequestCommand:
		c.sendTcpPacket(newTcpPacket(tcpCommand_HeartbeatResponseCommand, 0, p.CorrelationId, nil))
	case tcpCommand_SubscriptionConfirmation:
		ch := asyncResult.(chan *SubscriptionResult)
		r := &SubscriptionResult{
			conn:          c,
			correlationId: p.CorrelationId,
			Confirmation:  res.(*SubscriptionConfirmation),
			Events:        make(chan *StreamEventAppeared),
		}
		ch <- r
		close(ch)
		c.asyncResults[p.CorrelationId] = r
	case tcpCommand_StreamEventAppeared:
		r := asyncResult.(*SubscriptionResult)
		r.Events <- res.(*StreamEventAppeared)
	case tcpCommand_SubscriptionDropped:
		r := asyncResult.(*SubscriptionResult)
		r.unsubscribe <- res.(*SubscriptionDropped)
		close(r.Events)
		close(r.unsubscribe)
		delete(c.asyncResults, p.CorrelationId)
	}
}

func parsePayload(tc tcpCommand, payload []byte) (res proto.Message) {
	switch tc {
	case tcpCommand_WriteEventsCompleted:
		res = &WriteEventsCompleted{}
	case tcpCommand_ReadStreamEventsForwardCompleted:
		res = &ReadStreamEventsCompleted{}
	case tcpCommand_SubscriptionConfirmation:
		res = &SubscriptionConfirmation{}
	case tcpCommand_StreamEventAppeared:
		res = &StreamEventAppeared{}
	case tcpCommand_SubscriptionDropped:
		res = &SubscriptionDropped{}
	}
	if res != nil {
		proto.Unmarshal(payload, res)
	}
	return
}

func (c *connection) sendTcpPacket(p *tcpPacket) error {
    if err := binary.Write(c.conn, binary.LittleEndian, p.Size()); err != nil {
        return err
    }
	if _, err := c.conn.Write(p.Bytes()); err != nil {
		return err
	}
	log.Info("Sent Command: %s | CorrelationId: %s", p.Command, p.CorrelationId)
	return nil
}

func closeIfChan(ch interface{}) {
	v := reflect.ValueOf(ch)
	if v.Kind() == reflect.Chan {
		v.Close()
	}
}
