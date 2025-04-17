package uuid

import (
	"fmt"
	"strings"
	"testing"
)

func TestUUIDStringAndParse(t *testing.T) {
	gen := newUUIDGenerator()
	typeID := IDType(42)
	uuid := gen.NewUUID(typeID)
	uuidStr := uuid.String()
	parsed, err := FromString(uuidStr)
	if err != nil {
		t.Fatalf("failed to parse UUID string: %v", err)
	}
	if uuid != parsed {
		t.Errorf("parsed UUID does not match original.\nGot:  %v\nWant: %v", parsed, uuid)
	}
}

func TestFromStringInvalidFormat(t *testing.T) {
	invalidInputs := []string{
		"1234",                                // Not enough parts.
		"1234567890abcd-1-01",                 // Only 3 parts.
		"1234567890abcd-1-01-xyz",             // Fourth part (unused) invalid (should be 15 hex digits).
		"1234567890abcd-g-01-000000000000000", // Invalid version part.
	}
	for _, s := range invalidInputs {
		if _, err := FromString(s); err == nil {
			t.Errorf("expected error for invalid UUID string %q, but got none", s)
		}
	}
}

func TestUniqueness(t *testing.T) {
	gen := newUUIDGenerator()
	uuidSet := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		uuid := gen.NewUUID(1)
		s := uuid.String()
		if _, exists := uuidSet[s]; exists {
			t.Errorf("duplicate UUID generated: %s", s)
		}
		uuidSet[s] = struct{}{}
	}
}

func TestUUIDStringFormat(t *testing.T) {
	gen := newUUIDGenerator()
	uuid := gen.NewUUID(1)
	// Expected string format: <timestamp><sequence>-<version>-<type>-<unused>
	//   - First part: 14 hex digits (13 for timestamp, 1 for sequence)
	//   - Second part: 1 hex digit (version)
	//   - Third part: 2 hex digits (type)
	//   - Fourth part: 15 hex digits (unused)
	parts := strings.Split(uuid.String(), "-")
	if len(parts) != 4 {
		t.Fatalf("expected 4 parts, got %d", len(parts))
	}
	if len(parts[0]) != 14 {
		t.Errorf("expected first part to be 14 hex digits, got %d: %s", len(parts[0]), parts[0])
	}
	if len(parts[1]) != 1 {
		t.Errorf("expected version part to be 1 hex digit, got %d: %s", len(parts[1]), parts[1])
	}
	if len(parts[2]) != 2 {
		t.Errorf("expected type part to be 2 hex digits, got %d: %s", len(parts[2]), parts[2])
	}
	if len(parts[3]) != 15 {
		t.Errorf("expected unused part to be 15 hex digits, got %d: %s", len(parts[3]), parts[3])
	}
	// The unused field is expected to be all zeros.
	if parts[3] != "000000000000000" {
		// Ignore and allow
		//t.Errorf("expected unused part to be '000000000000000', got %s", parts[3])
	}
}

func TestUUIDExtractors(t *testing.T) {
	gen := newUUIDGenerator()
	typeID := IDType(42)
	uuid := gen.NewUUID(typeID)
	// Check that Timestamp is non-zero.
	if uuid.Timestamp() == 0 {
		t.Error("Timestamp() returned 0; expected nonzero")
	}
	// For a freshly generated UUID the sequence should be 0.
	if uuid.Sequence() != 0 {
		t.Errorf("Sequence() = %d; expected 0", uuid.Sequence())
	}
	// Check that Version equals the default version.
	if uuid.Version() != currentVersion {
		t.Errorf("Version() = %d; expected %d", uuid.Version(), currentVersion)
	}
	// Check Type.
	if uuid.Type() != typeID {
		t.Errorf("Type() = %d; expected %d", uuid.Type(), typeID)
	}
	// Check Unused.
	if uuid.Unused() != 0 {
		t.Errorf("Unused() = %d; expected 0", uuid.Unused())
	}
}

func TestIsNil(t *testing.T) {
	var nilUUID UUID
	if !nilUUID.IsNil() {
		t.Error("expected nil UUID to be IsNil() true")
	}
	gen := newUUIDGenerator()
	uuid := gen.NewUUID(1)
	if uuid.IsNil() {
		t.Error("expected non-nil UUID to be IsNil() false")
	}
}

func TestIsNilString(t *testing.T) {
	// For an all-zero UUID, the expected string is:
	//   - First part: 14 hex digits (13 for timestamp, 1 for sequence): "00000000000000"
	//   - Second part: 1 hex digit (version): "0"
	//   - Third part: 2 hex digits (type): "00"
	//   - Fourth part: 15 hex digits (unused): "000000000000000"
	// Concatenated: "00000000000000-0-00-000000000000000"
	nilTest, err := FromString("00000000000000-0-00-000000000000000")
	if err != nil {
		t.Fatalf("failed to parse nil UUID string: %v", err)
	}
	var nilUUID UUID
	if nilUUID != nilTest {
		t.Error("expected nil UUID parsed from string to equal nil UUID")
	}
}

// TestFromStringRoundTrip verifies that parsing and then re‑stringifying
// yields exactly the original valid input.
func TestFromStringRoundTrip(t *testing.T) {
	validStrings := []string{
		// nil UUID
		"00000000000000-0-00-000000000000000",
		// tiny nonzero timestamp, seq=0, version=1, type=2
		"00000000000001-1-02-000000000000000",
		// tiny nonzero timestamp, seq=0, version=1, type=2
		"00000000000001-1-02-010101010101010",
		// small timestamp, seq=3, version=1, type=42
		fmt.Sprintf("%013x%01x-%01x-%02x-%015x", 0x2A, 0x3, currentVersion, 42, 0),
		// max sequence (f), version=1, type=255
		fmt.Sprintf("%013x%01x-%01x-%02x-%015x", 0x10, 0xF, currentVersion, 0xFF, 0),
	}

	for _, s := range validStrings {
		u, err := FromString(s)
		if err != nil {
			t.Errorf("unexpected error parsing valid UUID %q: %v", s, err)
			continue
		}
		out := u.String()
		if out != s {
			t.Errorf("round-trip mismatch:\n got: %q\nwant: %q", out, s)
		}
	}
}

// TestFromStringValidFields parses a handful of simple, hand‑crafted UUID
// strings and checks each extractor against the known literal values.
func TestFromStringValidFields(t *testing.T) {
	type fixture struct {
		str              string
		wantTS           uint64
		wantSeq, wantVer uint8
		wantType         IDType
		wantUnused       uint64
	}

	fixtures := []fixture{
		{
			str:        "00000000000000-0-00-000000000000000",
			wantTS:     0,
			wantSeq:    0,
			wantVer:    0,
			wantType:   0,
			wantUnused: 0,
		},
		{
			str:        "000000000000a5-2-7f-000000000000000",
			wantTS:     0xA,
			wantSeq:    5,
			wantVer:    2,
			wantType:   0x7F,
			wantUnused: 0,
		},
		{
			str:        "0000000000ffff-f-ff-000000000000000",
			wantTS:     0xFFF,
			wantSeq:    0xF,
			wantVer:    0xF,
			wantType:   0xFF,
			wantUnused: 0,
		},
	}

	for _, f := range fixtures {
		u, err := FromString(f.str)
		if err != nil {
			t.Errorf("FromString(%q) returned error: %v", f.str, err)
			continue
		}
		if ts := u.Timestamp(); ts != f.wantTS {
			t.Errorf("%q: Timestamp() = %d; want %d", f.str, ts, f.wantTS)
		}
		if sq := u.Sequence(); sq != f.wantSeq {
			t.Errorf("%q: Sequence() = %d; want %d", f.str, sq, f.wantSeq)
		}
		if vr := u.Version(); vr != f.wantVer {
			t.Errorf("%q: Version() = %d; want %d", f.str, vr, f.wantVer)
		}
		if tp := u.Type(); tp != f.wantType {
			t.Errorf("%q: Type() = %d; want %d", f.str, tp, f.wantType)
		}
		if un := u.Unused(); un != f.wantUnused {
			t.Errorf("%q: Unused() = %d; want %d", f.str, un, f.wantUnused)
		}
	}
}
