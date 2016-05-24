package gesclient

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"io"
	"net"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Connection interface {
	WaitForConnection()
	Close() error
	AppendToStream(stream string, expectedVersion int, events []*EventData,
		userCredentials *UserCredentials) (*WriteResult, error)
	AppendToStreamAsync(stream string, expectedVersion int, events []*EventData,
		userCredentials *UserCredentials) (<-chan *WriteResult, error)
	ReadStreamEventsForward(stream string, start int, max int,
		userCredentials *UserCredentials) (*StreamEventsSlice, error)
	ReadStreamEventsForwardAsync(stream string, start int, max int,
		userCredentials *UserCredentials) (<-chan *StreamEventsSlice, error)
	SubscribeToStream(stream string, userCredentials *UserCredentials) (Subscription, error)
	SubscribeToStreamAsync(stream string, userCredentials *UserCredentials) (<-chan Subscription, error)
}

type connection struct {
	address     string
	opMutex     sync.RWMutex
	operations  map[uuid.UUID]operation
	leftOver    []byte
	reconnect   *int32
	connected   *int32
	output      chan *tcpPacket
	conn        net.Conn
	readerEnded chan struct{}
	writerEnded chan struct{}
}

func NewConnection(addr string) Connection {
	reconnect := int32(1)
	connected := int32(0)
	c := &connection{
		address:     addr,
		operations:  make(map[uuid.UUID]operation),
		output:      make(chan *tcpPacket, 100),
		reconnect:   &reconnect,
		connected:   &connected,
		readerEnded: make(chan struct{}),
		writerEnded: make(chan struct{}),
	}
	go c.connect()
	return c
}

func (c *connection) connect() {
	var err error
	for atomic.LoadInt32(c.reconnect) == 1 {
		log.Info("Connecting to %s", c.address)
		c.conn, err = net.DialTimeout("tcp4", c.address, time.Second*3)
		if err == nil {
			go c.reader()
			go c.writer()

			atomic.StoreInt32(c.connected, 1)
			c.resubscribe()

			<-c.writerEnded

			atomic.StoreInt32(c.connected, 0)
			log.Info("Disconnected from %s", c.address)
		} else {
			log.Error("Connection failed: %v", err)
		}
		if atomic.LoadInt32(c.reconnect) == 1 {
			time.Sleep(time.Second * 3)
		}
	}
	close(c.readerEnded)
	close(c.writerEnded)
	c.clearOperations()
}

func (c *connection) reader() {
	log.Info("Starting reader")
	for {
		buff := make([]byte, 4096)
		n, err := c.conn.Read(buff)
		if err == nil {
			c.onData(buff[:n])
		} else if err == io.EOF {
			break
		} else {
			if isClosedConnError(err) {
				break
			}
			log.Error("conn.Read: %v", err)
		}
	}
	c.readerEnded <- struct{}{}
	log.Info("Reader ended")
}

func (c *connection) writer() {
	log.Info("Starting writer")
	var packet *tcpPacket
	run := true
	for run {
		if packet == nil {
			select {
			case packet = <-c.output:
			case <-c.readerEnded:
				run = false
				continue
			default:
				continue
			}
		}
		if err := c.writeToConnection(c.conn, packet); err == nil {
			packet = nil
		} else if err == io.EOF {
			break
		}
	}
	c.writerEnded <- struct{}{}
	log.Info("Writer ended")
}

func (c *connection) writeToConnection(conn net.Conn, packet *tcpPacket) error {
	if err := binary.Write(conn, binary.LittleEndian, packet.Size()); err != nil {
		log.Error("binary.Write failed: %v", err)
		return err
	}
	if _, err := conn.Write(packet.Bytes()); err != nil {
		log.Error("net.Conn.Write failed: %v", err)
		return err
	}
	log.Debug("Sent Command: %s | CorrelationId: %s", packet.Command, packet.CorrelationId)
	return nil
}

func (c *connection) assertConnected() error {
	if atomic.LoadInt32(c.connected) == 0 {
		return errors.New("Not connected")
	}
	return nil
}

func (c *connection) WaitForConnection() {
	for atomic.LoadInt32(c.connected) == 0 {
		time.Sleep(10)
	}
}

func (c *connection) Close() error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	atomic.StoreInt32(c.reconnect, 0)
	return c.conn.Close()
}

func (c *connection) AppendToStream(
	stream string,
	expectedVersion int,
	events []*EventData,
	userCredentials *UserCredentials,
) (*WriteResult, error) {
	ch, err := c.AppendToStreamAsync(stream, expectedVersion, events, userCredentials)
	if err != nil {
		return nil, err
	}
	return <-ch, err
}

func (c *connection) AppendToStreamAsync(
	stream string,
	expectedVersion int,
	events []*EventData,
	userCredentials *UserCredentials,
) (<-chan *WriteResult, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	op := newAppendToStreamOperation(stream, events, expectedVersion, userCredentials)
	return op.GetResultChannel(), c.enqueueOperation(op, true)
}

func (c *connection) ReadStreamEventsForward(
	stream string,
	start int,
	max int,
	userCredentials *UserCredentials,
) (*StreamEventsSlice, error) {
	ch, err := c.ReadStreamEventsForwardAsync(stream, start, max, userCredentials)
	if err != nil {
		return nil, err
	}
	return <-ch, nil
}

func (c *connection) ReadStreamEventsForwardAsync(
	stream string,
	start int,
	max int,
	userCredentials *UserCredentials,
) (<-chan *StreamEventsSlice, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	op := newReadStreamEventsForwardOperation(stream, start, max, userCredentials)
	return op.GetResultChannel(), c.enqueueOperation(op, true)
}

func (c *connection) SubscribeToStream(stream string, userCredentials *UserCredentials) (Subscription, error) {
	ch, err := c.SubscribeToStreamAsync(stream, userCredentials)
	if err != nil {
		return nil, err
	}
	return <-ch, nil
}

func (c *connection) SubscribeToStreamAsync(stream string, userCredentials *UserCredentials) (<-chan Subscription, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	op := newSubscribeToStreamOperation(stream, c, userCredentials)
	return op.GetResultChannel(), c.enqueueOperation(op, true)
}

func (c *connection) enqueueOperation(op operation, isNew bool) error {
	payload, err := proto.Marshal(op.GetRequestMessage())
	if err != nil {
		log.Error("Sending command failed: %v", err)
		op.Fail(fmt.Errorf("Sending command failed: %v", err))
		return err
	}

	correlationId := op.GetCorrelationId()
	userCredentials := op.UserCredentials()
	var authFlag byte = 0
	if userCredentials != nil {
		authFlag = 1
	}
	c.output <- newTcpPacket(
		op.GetRequestCommand(),
		authFlag,
		correlationId,
		payload,
		userCredentials,
	)

	if isNew {
		c.addOperation(correlationId, op)
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
	log.Debug("Received Command: %s | CorrelationId: %s", p.Command, p.CorrelationId)

	operation := c.getOperation(p.CorrelationId)

	if operation != nil {
		operation.ParseResponse(p)
		if operation.IsCompleted() {
			c.removeOperation(p.CorrelationId)
		} else if operation.Retry() {
			c.enqueueOperation(operation, false)
		}
		return
	}

	switch p.Command {
	case tcpCommand_HeartbeatRequestCommand:
		c.output <- newTcpPacket(tcpCommand_HeartbeatResponseCommand, 0, p.CorrelationId, nil, nil)
	default:
		log.Error("Command not supported")
	}
}

func (c *connection) addOperation(correlationId uuid.UUID, operation operation) {
	c.opMutex.Lock()
	c.operations[correlationId] = operation
	c.opMutex.Unlock()
}

func (c *connection) getOperation(correlationId uuid.UUID) operation {
	c.opMutex.RLock()
	operation := c.operations[correlationId]
	c.opMutex.RUnlock()
	return operation
}

func (c *connection) removeOperation(correlationId uuid.UUID) {
	c.opMutex.Lock()
	delete(c.operations, correlationId)
	c.opMutex.Unlock()
}

func (c *connection) resubscribe() {
	c.opMutex.RLock()
	for _, op := range c.operations {
		c.enqueueOperation(op, false)
	}
	c.opMutex.RUnlock()
}

func (c *connection) clearOperations() {
	c.opMutex.RLock()
	for id, op := range c.operations {
		op.Fail(errors.New("Connection closed"))
		delete(c.operations, id)
	}
	c.opMutex.RUnlock()
}

// Copied from http2\server
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	// TODO: remove this string search and be more like the Windows
	// case below. That might involve modifying the standard library
	// to return better error types.
	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}

	// TODO(bradfitz): x/tools/cmd/bundle doesn't really support
	// build tags, so I can't make an http2_windows.go file with
	// Windows-specific stuff. Fix that and move this, once we
	// have a way to bundle this into std's net/http somehow.
	if runtime.GOOS == "windows" {
		if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
			if se, ok := oe.Err.(*os.SyscallError); ok && se.Syscall == "wsarecv" {
				const WSAECONNABORTED = 10053
				const WSAECONNRESET = 10054
				if n := errno(se.Err); n == WSAECONNRESET || n == WSAECONNABORTED {
					return true
				}
			}
		}
	}
	return false
}

func errno(v error) uintptr {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Uintptr {
		return uintptr(rv.Uint())
	}
	return 0
}
