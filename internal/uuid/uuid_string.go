package uuid

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	ErrInvalidLength  = errors.New(`invalid UUID length`)
	ErrInvalidVersion = errors.New(`invalid UUID version`)
)

// Checks version, passes off to version specific parser
func FromString(s string) (UUID, error) {

	strLen := len(s)

	// empty string is a nil UUID
	// 0 version reserved for nil
	if strLen < 1 || s[0:1] == `0` {
		return NilUUID, nil
	}

	vUint, err := strconv.ParseUint(s[0:1], 16, 8)
	if err != nil || vUint > math.MaxUint8 {
		return UUID{}, ErrInvalidVersion
	}

	var version byte = byte(vUint)

	if version == 1 {
		return fromString_v1(s)
	}

	return UUID{}, fmt.Errorf("invalid UUID version: %d", version)
}

func toString_v1(u UUID) string {
	ts := u.Timestamp()
	seq := u.Sequence()
	typ := u.Type()
	un := u.Unused()
	// 1-000000000000000-00-00000000000000
	return fmt.Sprintf("1-%013x%02x-%02x-%014x", ts, seq, typ, un)
}

// 1-000000000000000-00-00000000000000
// <version:1 hex digit>-<timestamp:13 hex digits><sequence:2 hex digits>-<type:2 hex digits>-<unused:14 hex digits>
func fromString_v1(s string) (UUID, error) {

	if len(s) != 35 {
		return UUID{}, ErrInvalidLength
	}

	parts := strings.Split(s, "-")

	if len(parts) != 4 {
		return UUID{}, errors.New("invalid UUID format: must have four parts")
	}
	// parts[0]=version (1 hex)
	// parts[1]=timestamp+sequence (15 hex)
	// parts[2]=type (2 hex)
	// parts[3]=unused (14 hex)
	if len(parts[0]) != 1 {
		return UUID{}, errors.New("invalid version part length, expected 1 hex digit")
	}
	if len(parts[1]) != 15 {
		return UUID{}, errors.New("invalid timestamp+sequence part length, expected 15 hex digits")
	}
	if len(parts[2]) != 2 {
		return UUID{}, errors.New("invalid type part length, expected 2 hex digits")
	}
	if len(parts[3]) != unusedBits/4 {
		return UUID{}, fmt.Errorf("invalid unused part length, expected %d hex digits", unusedBits/4)
	}

	// version
	ver := uint64(1)

	// parse timestamp and sequence
	ts, err := strconv.ParseUint(parts[1][:13], 16, timestampBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	seq, err := strconv.ParseUint(parts[1][13:], 16, sequenceBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid sequence: %w", err)
	}
	// parse type
	typVal, err := strconv.ParseUint(parts[2], 16, typeBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid type: %w", err)
	}
	// parse "unused"
	unused, err := strconv.ParseUint(parts[3], 16, unusedBits)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid unused: %w", err)
	}

	// rebuild the two 64-bit halves
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
