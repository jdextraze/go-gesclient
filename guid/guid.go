package guid

import u "github.com/satori/go.uuid"

func ToBytes(uuid u.UUID) []byte {
	bytes := uuid.Bytes()
	guid := make([]byte, 16)
	guid[3] = bytes[0]
	guid[2] = bytes[1]
	guid[1] = bytes[2]
	guid[0] = bytes[3]
	guid[5] = bytes[4]
	guid[4] = bytes[5]
	guid[7] = bytes[6]
	guid[6] = bytes[7]
	copy(guid[8:], bytes[8:])
	return guid
}

func FromBytes(guid []byte) u.UUID {
	var bytes [16]byte
	bytes[0] = guid[3]
	bytes[1] = guid[2]
	bytes[2] = guid[1]
	bytes[3] = guid[0]
	bytes[4] = guid[5]
	bytes[5] = guid[4]
	bytes[6] = guid[7]
	bytes[7] = guid[6]
	copy(bytes[8:], guid[8:])
	return u.UUID(bytes)
}
