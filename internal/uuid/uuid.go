package uuid

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	// NilUUID is the zero-valued UUID.
	NilUUID = UUID{}
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
	nanoseconds := (t % 1e6) * 1000
	return time.Unix(seconds, nanoseconds)
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
	return u == NilUUID
}

// String returns the UUID as a formatted string.
// "<timestamp:13 hex digits><sequence:2 hex digits>-<version:1 hex digit>-<type:2 hex digits>-<unused:14 hex digits>"
func (u UUID) String() string {
	ts := u.Timestamp()
	seq := u.Sequence()
	ver := u.Version()
	typ := u.Type()
	un := u.Unused()
	return fmt.Sprintf("%013x%02x-%01x-%02x-%014x", ts, seq, ver, typ, un)
}

// FromString parses a UUID from its string representation.
func FromString(s string) (UUID, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 4 {
		return UUID{}, errors.New("invalid UUID format: must have four parts")
	}
	if len(parts[0]) != 15 {
		return UUID{}, errors.New("invalid first part length, expected 15 hex digits")
	}
	if len(parts[1]) != 1 {
		return UUID{}, errors.New("invalid version part length, expected 1 hex digit")
	}
	if len(parts[2]) != 2 {
		return UUID{}, errors.New("invalid type part length, expected 2 hex digits")
	}
	if len(parts[3]) != unusedBits/4 {
		return UUID{}, errors.New("invalid unused part length, expected 14 hex digits")
	}

	ts, err := strconv.ParseUint(parts[0][:13], 16, timestampBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	seq, err := strconv.ParseUint(parts[0][13:], 16, sequenceBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid sequence: %w", err)
	}
	ver, err := strconv.ParseUint(parts[1], 16, versionBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid version: %w", err)
	}
	typVal, err := strconv.ParseUint(parts[2], 16, typeBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid type: %w", err)
	}
	unused, err := strconv.ParseUint(parts[3], 16, unusedBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid unused: %w", err)
	}

	typeHigh := (typVal >> (typeBits / 2)) & ((1 << (typeBits / 2)) - 1)
	high := (ts << (sequenceBits + versionBits + (typeBits / 2))) |
		(seq << (versionBits + (typeBits / 2))) |
		(ver << (typeBits / 2)) |
		typeHigh

	typeLow := typVal & ((1 << (typeBits / 2)) - 1)
	low := (typeLow << unusedBits) | unused

	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], high)
	binary.BigEndian.PutUint64(uuid[8:16], low)
	return uuid, nil
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

// NewUUID generates a new UUID for the given typeID (0 < typeID < 256).
func (g *UUIDGenerator) NewUUID(typeID IDType) UUID {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := uint64(time.Now().UnixMicro()) - customEpoch
	if now == g.lastTimestamp {
		g.sequence++
		if g.sequence >= (1 << sequenceBits) {
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

// New creates a new UUID with optional typeID.
func New(typeId ...IDType) UUID {
	if len(typeId) > 0 {
		return generator.NewUUID(typeId[0])
	}
	return generator.NewUUID(0)
}
