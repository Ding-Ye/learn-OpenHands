package eventstore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// UUID is a 128-bit identifier rendered as the canonical 8-4-4-4-12 hex
// string. We roll our own because s01 deliberately ships zero
// dependencies — every learner can audit the entire module in one
// `go doc` pass.
type UUID [16]byte

// NewUUID generates an RFC 4122 v4 UUID (random, with the version and
// variant bits set per spec). It panics if the OS RNG fails, since at
// that point the program cannot do anything meaningful.
func NewUUID() UUID {
	var u UUID
	if _, err := rand.Read(u[:]); err != nil {
		panic(fmt.Sprintf("rand.Read: %v", err))
	}
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant RFC 4122
	return u
}

func (u UUID) String() string {
	var buf [36]byte
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], u[10:16])
	return string(buf[:])
}

// MarshalJSON / UnmarshalJSON make UUIDs round-trip as plain strings.
func (u UUID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}

func (u *UUID) UnmarshalJSON(b []byte) error {
	if len(b) != 38 || b[0] != '"' || b[37] != '"' {
		return fmt.Errorf("uuid: bad length %d", len(b))
	}
	return u.parse(string(b[1:37]))
}

func (u *UUID) parse(s string) error {
	if len(s) != 36 {
		return fmt.Errorf("uuid: bad length %d", len(s))
	}
	for _, i := range []int{8, 13, 18, 23} {
		if s[i] != '-' {
			return fmt.Errorf("uuid: expected '-' at %d", i)
		}
	}
	hexes := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	raw, err := hex.DecodeString(hexes)
	if err != nil {
		return fmt.Errorf("uuid: %w", err)
	}
	copy(u[:], raw)
	return nil
}

// ParseUUID is the inverse of String. Useful in tests and CLI demos.
func ParseUUID(s string) (UUID, error) {
	var u UUID
	err := u.parse(s)
	return u, err
}
