package client

import (
	"errors"
	"fmt"
	"github.com/jdextraze/go-gesclient/guid"
	"github.com/satori/go.uuid"
)

const (
	PackageCommandOffset     = 0
	PackageFlagsOffset       = PackageCommandOffset + 1
	PackageCorrelationOffset = PackageFlagsOffset + 1
	PackageAuthOffset        = PackageCorrelationOffset + 16
	PackageMandatorySize     = PackageAuthOffset
)

type TcpFlag byte

const (
	FlagsNone          TcpFlag = 0x00
	FlagsAuthenticated TcpFlag = 0x01
)

var tcpFlags = map[byte]string{
	0x00: "None",
	0x01: "Authenticated",
}

func (f TcpFlag) String() string {
	return tcpFlags[byte(f)]
}

type Package struct {
	command       Command
	flags         TcpFlag
	correlationId uuid.UUID
	data          []byte
	username      string
	password      string
}

func NewTcpPackage(
	cmd Command,
	flags TcpFlag,
	correlationId uuid.UUID,
	data []byte,
	userCredentials *UserCredentials,
) *Package {
	var (
		username string
		password string
	)
	if flags&FlagsAuthenticated != 0 {
		if userCredentials == nil {
			panic("userCredentials are missing")
		}
		username = userCredentials.Username()
		if username == "" {
			panic("username is missing")
		}
		password = userCredentials.Password()
		if password == "" {
			panic("password is missing")
		}
	} else if userCredentials != nil {
		panic("userCredentials provided for non-authorized TcpPackage.")
	}
	return &Package{cmd, flags, correlationId, data, username, password}
}

func TcpPacketFromBytes(data []byte) (*Package, error) {
	dataLength := len(data)
	if dataLength < PackageMandatorySize {
		return nil, fmt.Errorf("data too short, length: %d", dataLength)
	}

	command := Command(data[PackageCommandOffset])
	flags := TcpFlag(data[PackageFlagsOffset])

	correlationId := guid.FromBytes(data[PackageCorrelationOffset:PackageAuthOffset])

	var (
		headerSize int = PackageMandatorySize
		username   string
		password   string
	)
	if flags&FlagsAuthenticated != 0 {
		usernameLength := int(data[PackageAuthOffset])
		usernameStartOffset := PackageAuthOffset + 1
		usernameEndOffset := usernameStartOffset + usernameLength
		if usernameEndOffset > dataLength {
			return nil, errors.New("Username length is too big, it does not fit into TcpPackage.")
		}
		username = string(data[usernameStartOffset:usernameEndOffset])

		passwordLength := int(data[usernameEndOffset])
		if usernameEndOffset+1+passwordLength > dataLength {
			return nil, errors.New("Password length is too big, it does not fit into TcpPackage.")
		}
		username = string(data[usernameEndOffset+1 : usernameEndOffset+1+passwordLength])

		headerSize += 1 + usernameLength + 1 + passwordLength
	}

	payload := data[headerSize:]

	return &Package{command, flags, correlationId, payload, username, password}, nil
}

func (p *Package) Bytes() []byte {
	contentLength := p.Size()
	bytes := make([]byte, contentLength)
	bytes[PackageCommandOffset] = byte(p.command)
	bytes[PackageFlagsOffset] = byte(p.flags)
	copy(bytes[PackageCorrelationOffset:], guid.ToBytes(p.correlationId))
	pos := PackageAuthOffset
	if p.flags&FlagsAuthenticated != 0 {
		bytes[pos] = byte(len(p.username))
		pos++
		copy(bytes[pos:], p.username)
		pos += len(p.username)
		bytes[pos] = byte(len(p.password))
		pos++
		copy(bytes[pos:], p.password)
		pos += len(p.password)
	}
	copy(bytes[pos:], p.data)
	return bytes
}

func (p *Package) Size() int32 {
	authLen := 0
	if p.flags&FlagsAuthenticated != 0 {
		authLen = 1 + len(p.username) + 1 + len(p.password)
	}
	return int32(PackageMandatorySize + authLen + len(p.data))
}

func (p *Package) Command() Command { return p.command }

func (p *Package) Flags() TcpFlag { return p.flags }

func (p *Package) CorrelationId() uuid.UUID { return p.correlationId }

func (p *Package) Username() string { return p.username }

func (p *Package) Password() string { return p.password }

func (p *Package) Data() []byte { return p.data }

func (p *Package) String() string {
	return fmt.Sprintf(
		"&{command:%s flags:%s correlationId:%s username:%s password:%s data:[...]}",
		p.command, p.flags, p.correlationId, p.username, p.password,
	)
}
