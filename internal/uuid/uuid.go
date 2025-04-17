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
	sequenceBits  = 4  // Sequence counter: allows 16 per microsecond.
	versionBits   = 4  // Version field.
	typeBits      = 8  // Type identifier.
	unusedBits    = 60 // Unused bits.
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

// Bit layout (overall bit positions 127..0):
// • Bits 127–76: Timestamp (52 bits)
// • Bits 75–72: Sequence (4 bits)
// • Bits 71–68: Version (4 bits)
// • Bits 67–64: Type high nibble (4 bits)
// • Bits 63–60: Type low nibble (4 bits)
// • Bits 59–0:  Unused (60 bits)

// Get the time UUID was generated
func (u UUID) Time() time.Time {
	t := int64(u.Timestamp() + customEpoch)
	seconds := t / 1e6
	nanoseconds := (t % 1e6) * 1000 // convert remaining microseconds to nanoseconds
	return time.Unix(seconds, nanoseconds)
}

// Timestamp returns the 52-bit timestamp (microseconds since Jan 1, 2025)
// extracted from the high 64 bits. (It occupies bits 127–76.)
func (u UUID) Timestamp() uint64 {
	high := binary.BigEndian.Uint64(u[0:8])
	return high >> 12 // Discard the lower 12 bits (sequence, version, typeHigh).
}

// Sequence returns the 4-bit sequence counter stored in bits 75–72.
func (u UUID) Sequence() uint8 {
	high := binary.BigEndian.Uint64(u[0:8])
	return uint8((high >> 8) & 0xF) // Extract bits 11..8 of the high 64-bit word.
}

// Version returns the 4-bit version stored in bits 71–68.
func (u UUID) Version() uint8 {
	high := binary.BigEndian.Uint64(u[0:8])
	return uint8((high >> 4) & 0xF) // Extract bits 7..4.
}

// Type returns the 8-bit type identifier, which is split across the high and low 64-bit words.
// • Type high nibble: bits 67–64 (stored as the lower 4 bits of the high word).
// • Type low nibble:  bits 63–60 (stored as the upper 4 bits of the low word).
func (u UUID) Type() IDType {
	high := binary.BigEndian.Uint64(u[0:8])
	low := binary.BigEndian.Uint64(u[8:16])
	typeHigh := high & 0xF                   // lower 4 bits of high word.
	typeLow := (low >> 60) & 0xF             // top 4 bits of low word.
	return IDType((typeHigh << 4) | typeLow) // Combine into one 8-bit value.
}

// Unused returns the 60-bit unused field from the low 64 bits (bits 59–0).
func (u UUID) Unused() uint64 {
	low := binary.BigEndian.Uint64(u[8:16])
	return low & ((uint64(1) << 60) - 1)
}

// IsNil indicates whether the UUID is the zero value.
func (u UUID) IsNil() bool {
	var z UUID
	return u == z
}

// String returns the UUID as a formatted string:
// "<timestamp:13 hex digits><sequence:1 hex digit>-<version:1 hex digit>-<type:2 hex digits>-<unused:15 hex digits>"
func (u UUID) String() string {
	ts := u.Timestamp() // 52 bits → 13 hex digits.
	seq := u.Sequence() // 4 bits → 1 hex digit.
	ver := u.Version()  // 4 bits → 1 hex digit.
	typ := u.Type()     // 8 bits → 2 hex digits.
	un := u.Unused()    // 60 bits → 15 hex digits.
	return fmt.Sprintf("%013x%01x-%01x-%02x-%015x", ts, seq, ver, typ, un)
}

// FromString parses a UUID from its string representation.
func FromString(s string) (UUID, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 4 {
		return UUID{}, errors.New("invalid UUID format: must have four parts")
	}
	if len(parts[0]) != 14 {
		return UUID{}, errors.New("invalid first part length, expected 14 hex digits")
	}
	if len(parts[1]) != 1 {
		return UUID{}, errors.New("invalid version part length, expected 1 hex digit")
	}
	if len(parts[2]) != 2 {
		return UUID{}, errors.New("invalid type part length, expected 2 hex digits")
	}
	if len(parts[3]) != 15 {
		return UUID{}, errors.New("invalid unused part length, expected 15 hex digits")
	}

	// Parse timestamp (first 13 hex digits) and sequence (last hex digit of first part).
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
	// Parse type (2 hex digits → 8 bits).
	typ, err := strconv.ParseUint(parts[2], 16, typeBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid type: %w", err)
	}

	// Parse type (2 hex digits → 8 bits).
	unused, err := strconv.ParseUint(parts[3], 16, unusedBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid type: %w", err)
	}

	// Pack fields into the two 64-bit words.
	// High 64 bits:
	// • timestamp: occupies the top 52 bits → shift left by 12.
	// • sequence: occupies the next 4 bits → shift left by 8.
	// • version: occupies the next 4 bits → shift left by 4.
	// • Type high nibble: lower 4 bits of type.
	typeHigh := (typ >> 4) & 0xF
	high := (ts << 12) | (seq << 8) | (ver << 4) | typeHigh

	// Low 64 bits:
	// • Type low nibble: lower 4 bits of type, placed in the upper 4 bits of the low word.
	// • Unused: occupies the lower 60 bits.
	typeLow := typ & 0xF
	low := (typeLow << unusedBits) | unused
	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], high)
	binary.BigEndian.PutUint64(uuid[8:16], low)
	return uuid, nil
}

// UUIDGenerator is responsible for generating unique UUIDs.
type UUIDGenerator struct {
	mu            sync.Mutex
	lastTimestamp uint64 // Last timestamp (delta from customEpoch).
	sequence      uint64
}

func newUUIDGenerator() UUIDGenerator {
	return UUIDGenerator{}
}

// NewUUID generates a new UUID for the given typeID (must be < 256).
func (g *UUIDGenerator) NewUUID(typeID IDType) UUID {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := uint64(time.Now().UnixMicro()) - customEpoch
	if now == g.lastTimestamp {
		g.sequence++
		if g.sequence >= (1 << sequenceBits) {
			// Busy-wait until the next microsecond.
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

	// Pack the fields.
	// High 64 bits: (timestamp << 12) | (sequence << 8) | (version << 4) | (type high nibble)
	typeHigh := (uint64(typeID) >> 4) & 0xF
	high := (now << 12) | (g.sequence << 8) | (uint64(currentVersion) << 4) | typeHigh

	// Low 64 bits: (type low nibble << 60) | unused (zero).
	typeLow := uint64(typeID) & 0xF
	low := (typeLow << unusedBits)
	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], high)
	binary.BigEndian.PutUint64(uuid[8:16], low)
	return uuid
}

// New creates a new UUID, optionally using a provided type identifier.
func New(typeId ...IDType) UUID {
	if len(typeId) > 0 {
		return generator.NewUUID(typeId[0])
	}
	return generator.NewUUID(0)
}
