package client

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"github.com/satori/go.uuid"
	"io"
	"net"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type PackageConnection struct {
	ipEndpoint            net.Addr
	connectionId          uuid.UUID
	ssl                   bool
	targetHost            string
	validateService       bool
	timeout               time.Duration
	packageHandler        func(conn *PackageConnection, packet *Package)
	errorHandler          func(conn *PackageConnection, err error)
	connectionEstablished func(conn *PackageConnection)
	connectionClosed      func(conn *PackageConnection, err error)
	conn                  net.Conn

	leftOver  []byte
	sendQueue chan *Package
	isClosed  int32

	localEndpoint net.Addr
}

func NewPackageConnection(
	ipEndpoint net.Addr,
	connectionId uuid.UUID,
	ssl bool,
	targetHost string,
	validateService bool,
	timeout time.Duration,
	packageHandler func(conn *PackageConnection, packet *Package),
	errorHandler func(conn *PackageConnection, err error),
	connectionEstablished func(conn *PackageConnection),
	connectionClosed func(conn *PackageConnection, err error),
) *PackageConnection {
	c := &PackageConnection{
		ipEndpoint:            ipEndpoint,
		connectionId:          connectionId,
		ssl:                   ssl,
		targetHost:            targetHost,
		validateService:       validateService,
		timeout:               timeout,
		packageHandler:        packageHandler,
		errorHandler:          errorHandler,
		connectionEstablished: connectionEstablished,
		connectionClosed:      connectionClosed,
		sendQueue:             make(chan *Package, 4096),
	}
	c.connect()
	return c
}

func (c *PackageConnection) ConnectionId() uuid.UUID { return c.connectionId }

func (c *PackageConnection) connect() {
	var conn net.Conn
	var err error

	if c.ssl {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: c.timeout}, c.ipEndpoint.Network(),
			c.ipEndpoint.String(), &tls.Config{})
	} else {
		conn, err = net.DialTimeout(c.ipEndpoint.Network(), c.ipEndpoint.String(), c.timeout)
	}

	if err != nil {
		log.Debugf("Connection to %s failed. Error: %v", c.ipEndpoint, err)
		if c.connectionClosed != nil {
			c.connectionClosed(c, err)
		}
	} else {
		c.localEndpoint = conn.LocalAddr()
		log.Debugf("Connection to %s succeeded.", c.ipEndpoint)
		c.conn = conn
		if c.connectionEstablished != nil {
			c.connectionEstablished(c)
		}
	}

	go c.sender()
}

func (c *PackageConnection) receiver() {
	var err error
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
			log.Errorf("conn.Read: %v", err)
		}
	}
	c.closeInternal("Socket receive error", err)
}

func (c *PackageConnection) sender() {
	var err error
	for p := range c.sendQueue {
		if err = binary.Write(c.conn, binary.LittleEndian, p.Size()); err != nil {
			log.Errorf("binary.Write failed: %v", err)
			break
		}
		if _, err = c.conn.Write(p.Bytes()); err != nil {
			log.Errorf("net.Conn.Write failed: %v", err)
			break
		}
		log.Debugf("Sent Command: %s | CorrelationId: %s", p.Command(), p.CorrelationId())
	}
	c.closeInternal("Socket send error.", err)
}

func (c *PackageConnection) closeInternal(reason string, socketError error) {
	if atomic.CompareAndSwapInt32(&c.isClosed, 0, 1) {
		close(c.sendQueue)
		log.Debugf("PackageConnection.closeInternal: %s. %v", reason, c.conn.Close())
		if c.connectionClosed != nil {
			c.connectionClosed(c, socketError)
		}
	}
}

const tcpPacketContentLengthSize = 4

func (c *PackageConnection) onData(data []byte) {
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
		p, _ := TcpPacketFromBytes(data[tcpPacketContentLengthSize:])
		c.packageHandler(c, p)
		//c.errorHandler(c, err)
	} else if dataLength > packetSize {
		c.onData(data[:packetSize])
		c.onData(data[packetSize:])
	} else {
		c.leftOver = data
	}
}

func (c *PackageConnection) RemoteEndpoint() net.Addr { return c.ipEndpoint }

func (c *PackageConnection) LocalEndpoint() net.Addr { return c.localEndpoint }

func (c *PackageConnection) StartReceiving() error {
	if c.conn == nil {
		return errors.New("Failed connection")
	}
	if c.IsClosed() {
		return errors.New("Connection is closed")
	}
	go c.receiver()
	return nil
}

func (c *PackageConnection) EnqueueSend(p *Package) error {
	if c.conn == nil {
		return errors.New("Failed connection")
	}
	if c.IsClosed() {
		return errors.New("Connection is closed")
	}
	c.sendQueue <- p
	return nil
}

func (c *PackageConnection) Close(reason string) error {
	if c.conn == nil {
		return errors.New("Failed connection")
	}
	if c.IsClosed() {
		return errors.New("Already closed")
	}

	return c.conn.Close()
}

func (c *PackageConnection) IsClosed() bool {
	return atomic.LoadInt32(&c.isClosed) == 1
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
