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

const (
	sequenceBits = 4 // 4-bit sequence counter: up to 16 UUIDs per nanosecond.
	versionBits  = 4 // 4-bit version field.
	typeBits     = 8 // 8-bit type identifier: supports up to 256 types.
	// Lower 64 bits total = 64 bits.
	// Unused bits = 64 - (sequenceBits + versionBits + typeBits)
	unusedBits = 64 - (sequenceBits + versionBits + typeBits) // 64 - (4+4+8)=48 bits.
)

const (
	currentVersion uint8  = 1              // The default version to use when generating a UUID.
	unusedString   string = "000000000000" // Expected constant for unused field (12 hex digits).
)

// UUID is a 128-bit custom identifier.
type UUID [16]byte

// Timestamp returns the nanosecond-precision timestamp stored in the upper 64 bits.
func (u UUID) Timestamp() uint64 {
	return binary.BigEndian.Uint64(u[0:8])
}

// Sequence extracts the 4-bit sequence counter from the lower 64 bits.
func (u UUID) Sequence() uint8 {
	lower := binary.BigEndian.Uint64(u[8:16])
	// Sequence occupies the top 4 bits (bits 63-60).
	return uint8((lower >> (versionBits + typeBits + unusedBits)) & ((1 << sequenceBits) - 1))
}

// Version extracts the 4-bit version field from the lower 64 bits.
func (u UUID) Version() uint8 {
	lower := binary.BigEndian.Uint64(u[8:16])
	// Version occupies the next 4 bits (bits 59-56).
	return uint8((lower >> (typeBits + unusedBits)) & ((1 << versionBits) - 1))
}

// Type extracts the 8-bit type identifier from the lower 64 bits.
func (u UUID) Type() uint8 {
	lower := binary.BigEndian.Uint64(u[8:16])
	// Type occupies the next 8 bits (bits 55-48).
	return uint8((lower >> unusedBits) & ((1 << typeBits) - 1))
}

// Unused extracts the 48-bit unused field from the lower 64 bits.
func (u UUID) Unused() uint64 {
	lower := binary.BigEndian.Uint64(u[8:16])
	return lower & ((uint64(1) << unusedBits) - 1)
}

// IsNil returns true if the UUID is entirely zero.
func (u UUID) IsNil() bool {
	var nilUUID UUID
	return u == nilUUID
}

// String returns a string representation of the UUID in the format:
// <timestamp><sequence>-<version>-<type>-<unused>
//   - timestamp: 16 hex digits (64 bits)
//   - sequence:  1 hex digit (4 bits) appended immediately after timestamp (total 17 hex digits)
//   - version:   1 hex digit (4 bits)
//   - type:      2 hex digits (8 bits)
//   - unused:    12 hex digits (48 bits), always "000000000000"
func (u UUID) String() string {
	ts := u.Timestamp()
	seq := u.Sequence()
	ver := u.Version()
	typ := u.Type()
	unused := u.Unused() // Should always be zero.
	return fmt.Sprintf("%016x%01x-%01x-%02x-%012x", ts, seq, ver, typ, unused)
}

// FromString converts a string representation of a UUID back into a UUID.
// Expected format: <timestamp><sequence>-<version>-<type>-<unused>
//   - First part: 17 hex digits (16 for timestamp, 1 for sequence)
//   - Second part: 1 hex digit (version)
//   - Third part: 2 hex digits (type)
//   - Fourth part: 12 hex digits (unused), which must equal "000000000000"
func FromString(s string) (UUID, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 4 {
		return UUID{}, errors.New("invalid UUID format: must have four hyphen-separated parts")
	}

	// First part: 17 hex digits.
	if len(parts[0]) != 17 {
		return UUID{}, errors.New("invalid first part length, expected 17 hex digits")
	}
	tsStr := parts[0][:16]
	seqStr := parts[0][16:]
	ts, err := strconv.ParseUint(tsStr, 16, 64)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	seq, err := strconv.ParseUint(seqStr, 16, sequenceBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid sequence: %w", err)
	}

	// Second part: version, 1 hex digit.
	if len(parts[1]) != 1 {
		return UUID{}, errors.New("invalid version part length, expected 1 hex digit")
	}
	ver, err := strconv.ParseUint(parts[1], 16, versionBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid version: %w", err)
	}

	// Third part: type, 2 hex digits.
	if len(parts[2]) != 2 {
		return UUID{}, errors.New("invalid type part length, expected 2 hex digits")
	}
	typ, err := strconv.ParseUint(parts[2], 16, typeBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid type: %w", err)
	}

	// Fourth part: unused, 12 hex digits. Instead of conversion, we directly check the string.
	if parts[3] != unusedString {
		return UUID{}, errors.New("unused field must be zero")
	}

	// Pack lower 64 bits: [ sequence (4 bits) | version (4 bits) | type (8 bits) | unused (48 bits) ]
	lower := (seq << (versionBits + typeBits + unusedBits)) |
		(uint64(ver) << (typeBits + unusedBits)) |
		(typ << unusedBits)
	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], ts)
	binary.BigEndian.PutUint64(uuid[8:16], lower)
	return uuid, nil
}

// UUIDGenerator generates new UUIDs while ensuring uniqueness.
type UUIDGenerator struct {
	mu            sync.Mutex
	lastTimestamp uint64
	sequence      uint64
}

// NewUUIDGenerator returns a new instance of UUIDGenerator.
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// NewUUID generates a new UUID using the provided typeID (must be < 256).
// It uses the current nanosecond timestamp and a sequence counter to ensure uniqueness,
// and sets the version field to the defaultVersion.
func (g *UUIDGenerator) NewUUID(tId ...uint8) (UUID, error) {
	var typeID uint8 = 0
	if len(tId) > 0 {
		typeID = tId[0]
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	now := uint64(time.Now().UnixNano())
	if now == g.lastTimestamp {
		g.sequence++
		if g.sequence >= (1 << sequenceBits) {
			// Busy-wait until the next nanosecond.
			for now <= g.lastTimestamp {
				now = uint64(time.Now().UnixNano())
			}
			g.sequence = 0
			g.lastTimestamp = now
		}
	} else {
		g.sequence = 0
		g.lastTimestamp = now
	}

	// Pack lower 64 bits: [ sequence (4 bits) | version (4 bits) | type (8 bits) | unused (48 bits) ]
	lower := (g.sequence << (versionBits + typeBits + unusedBits)) |
		(uint64(currentVersion) << (typeBits + unusedBits)) |
		(uint64(typeID) << unusedBits)
	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], now)
	binary.BigEndian.PutUint64(uuid[8:16], lower)
	return uuid, nil
}
