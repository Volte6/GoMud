package uuid

import (
	"encoding/binary"
	"sync"
	"time"
)

// Constants defining our field widths.
const (
	// customEpoch is January 1, 2025 in microseconds since the Unix epoch.
	customEpoch uint64 = 1735689600000000

	timestampBits = 52 // Timestamp: microseconds since customEpoch.
	sequenceBits  = 8  // Sequence counter: allows 256 per microsecond.
	versionBits   = 4  // Version field.
	typeBits      = 8  // Type identifier.
	unusedBits    = 56 // Unused bits.
)

var (
	// Pre-allocated nil value for comparison purposes
	nilUUID UUID = UUID{}
	// Generator singleton.
	generator = newUUIDGenerator()
	// currentVersion is the default version.
	currentVersion uint8 = 1
)

type IDType uint8

// UUID is a 128-bit identifier.
type UUID [16]byte

// Bit layout (big-endian bits 127..0):
//  • Bits 127–76: Timestamp (52 bits)
//  • Bits 75–68: Sequence (8 bits)
//  • Bits 67–64: Version (4 bits)
//  • Bits 63–60: Type high nibble (4 bits)
//  • Bits 59–56: Type low nibble (4 bits)
//  • Bits 55–0 : Unused (56 bits)

// Time returns the time the UUID was generated.
func (u UUID) Time() time.Time {
	t := int64(u.Timestamp() + customEpoch)
	seconds := t / 1e6
	nano := (t % 1e6) * 1000
	return time.Unix(seconds, nano)
}

// Timestamp returns the 52-bit timestamp (microseconds since Jan 1, 2025).
func (u UUID) Timestamp() uint64 {
	high := binary.BigEndian.Uint64(u[0:8])
	return high >> (sequenceBits + versionBits + (typeBits / 2))
}

// Sequence returns the 8-bit sequence counter.
func (u UUID) Sequence() uint8 {
	high := binary.BigEndian.Uint64(u[0:8])
	return uint8((high >> (versionBits + (typeBits / 2))) & ((1 << sequenceBits) - 1))
}

// Version returns the 4-bit version.
func (u UUID) Version() uint8 {
	high := binary.BigEndian.Uint64(u[0:8])
	return uint8((high >> (typeBits / 2)) & ((1 << versionBits) - 1))
}

// Type returns the 8-bit type identifier.
func (u UUID) Type() IDType {
	high := binary.BigEndian.Uint64(u[0:8])
	low := binary.BigEndian.Uint64(u[8:16])
	typeHigh := high & ((1 << (typeBits / 2)) - 1)
	typeLow := (low >> unusedBits) & ((1 << (typeBits / 2)) - 1)
	return IDType((typeHigh << (typeBits / 2)) | typeLow)
}

// Unused returns the 56-bit unused field.
func (u UUID) Unused() uint64 {
	low := binary.BigEndian.Uint64(u[8:16])
	return low & ((uint64(1) << unusedBits) - 1)
}

// IsNil indicates whether the UUID is the zero value.
func (u UUID) IsNil() bool {
	return u == nilUUID
}

// String returns the UUID as a formatted string.
func (u UUID) String() string {

	if u.IsNil() {
		return ``
	}

	// As new versions are incremented, the stringification may change as well
	if currentVersion == 1 {
		return toString_v1(u)
	}

	return ``
}

// MarshalText implements encoding.TextMarshaler.
func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *UUID) UnmarshalText(text []byte) error {
	parsed, err := FromString(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// UUIDGenerator is responsible for generating unique UUIDs.
type UUIDGenerator struct {
	mu            sync.Mutex
	lastTimestamp uint64
	sequence      uint64
}

func newUUIDGenerator() UUIDGenerator {
	return UUIDGenerator{}
}

// NewUUID generates a new UUID for the given typeID (0 ≤ typeID < 256).
func (g *UUIDGenerator) NewUUID(typeID IDType) UUID {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := uint64(time.Now().UnixMicro()) - customEpoch
	if now == g.lastTimestamp {
		g.sequence++
		if g.sequence >= (1 << sequenceBits) {
			// wait for next microsecond
			for {
				now = uint64(time.Now().UnixMicro()) - customEpoch
				if now > g.lastTimestamp {
					break
				}
			}
			g.sequence = 0
			g.lastTimestamp = now
		}
	} else {
		g.sequence = 0
		g.lastTimestamp = now
	}

	typeHigh := (uint64(typeID) >> (typeBits / 2)) & ((1 << (typeBits / 2)) - 1)
	high := (now << (sequenceBits + versionBits + (typeBits / 2))) |
		(g.sequence << (versionBits + (typeBits / 2))) |
		(uint64(currentVersion) << (typeBits / 2)) |
		typeHigh

	typeLow := uint64(typeID) & ((1 << (typeBits / 2)) - 1)
	low := (typeLow << unusedBits)

	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], high)
	binary.BigEndian.PutUint64(uuid[8:16], low)
	return uuid
}

// New creates a new UUID with an optional typeID.
func New(typeID ...IDType) UUID {
	if len(typeID) > 0 {
		return generator.NewUUID(typeID[0])
	}
	return generator.NewUUID(0)
}
