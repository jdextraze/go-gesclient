// Copyright (C) 2013-2018 by Maxim Bublis <b@codemonkey.ru>
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Copied from github.com/satori/go.uuid and modified to only support V4.

// Package uuid provides implementation of Universally Unique Identifier (UUID).
// Only support version 4 (as specified in RFC 4122).
package uuid

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
)

// UUID representation compliant with specification described in RFC 4122.
type UUID [16]byte

// Bytes returns bytes slice representation of UUID.
func (u UUID) Bytes() []byte {
	return u[:]
}

// Returns canonical string representation of UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

// Nil is special form of UUID that is specified to have all 128 bits set to zero.
var Nil = UUID{}

// Must is a helper that wraps a call to a function returning (UUID, error)
// and panics if the error is non-nil.
func Must(u UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return u
}

// NewV4 returns random generated UUID.
func NewV4() (UUID, error) {
	var u UUID
	if _, err := rand.Read(u[:]); err != nil {
		return Nil, err
	}
	// Version 4
	u[6] = (u[6] & 0x0f) | (4 << 4)
	// RFC4122 Variant
	u[8] = u[8]&(0xff>>2) | (0x02 << 6)
	return u, nil
}

// Equal returns true if u1 and u2 equals, otherwise returns false.
func Equal(u1 UUID, u2 UUID) bool {
	return bytes.Equal(u1[:], u2[:])
}
