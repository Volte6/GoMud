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
		"1234567890abcde-1-01",                // Only 3 parts.
		"1234567890abcde-1-01-xyz",            // Fourth part invalid length.
		"1234567890abcde-g-01-00000000000000", // Invalid version part.
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
	//   - First part: 15 hex digits (13 for timestamp, 2 for sequence)
	//   - Second part: 1 hex digit (version)
	//   - Third part: 2 hex digits (type)
	//   - Fourth part: 14 hex digits (unused)
	parts := strings.Split(uuid.String(), "-")
	if len(parts) != 4 {
		t.Fatalf("expected 4 parts, got %d", len(parts))
	}
	if len(parts[0]) != 15 {
		t.Errorf("expected first part to be 15 hex digits, got %d: %s", len(parts[0]), parts[0])
	}
	if len(parts[1]) != 1 {
		t.Errorf("expected version part to be 1 hex digit, got %d: %s", len(parts[1]), parts[1])
	}
	if len(parts[2]) != 2 {
		t.Errorf("expected type part to be 2 hex digits, got %d: %s", len(parts[2]), parts[2])
	}
	if len(parts[3]) != 14 {
		t.Errorf("expected unused part to be 14 hex digits, got %d: %s", len(parts[3]), parts[3])
	}
}

func TestUUIDExtractors(t *testing.T) {
	gen := newUUIDGenerator()
	typeID := IDType(42)
	uuid := gen.NewUUID(typeID)
	if uuid.Timestamp() == 0 {
		t.Error("Timestamp() returned 0; expected nonzero")
	}
	if uuid.Sequence() != 0 {
		t.Errorf("Sequence() = %d; expected 0", uuid.Sequence())
	}
	if uuid.Version() != currentVersion {
		t.Errorf("Version() = %d; expected %d", uuid.Version(), currentVersion)
	}
	if uuid.Type() != typeID {
		t.Errorf("Type() = %d; expected %d", uuid.Type(), typeID)
	}
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
	nilTest, err := FromString("000000000000000-0-00-00000000000000")
	if err != nil {
		t.Fatalf("failed to parse nil UUID string: %v", err)
	}
	var nilUUID UUID
	if nilUUID != nilTest {
		t.Error("expected nil UUID parsed from string to equal nil UUID")
	}
}

func TestFromStringRoundTrip(t *testing.T) {
	validStrings := []string{
		"000000000000000-0-00-00000000000000",
		"000000000000100-1-02-00000000000000",
		"000000000000100-1-02-01010101010101",
		fmt.Sprintf("%013x%02x-%01x-%02x-%014x", 0x2A, 0x03, currentVersion, 42, 0),
		fmt.Sprintf("%013x%02x-%01x-%02x-%014x", 0x10, 0xFF, currentVersion, 0xFF, 0),
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

func TestFromStringValidFields(t *testing.T) {
	type fixture struct {
		str        string
		wantTS     uint64
		wantSeq    uint8
		wantVer    uint8
		wantType   IDType
		wantUnused uint64
	}
	fixtures := []fixture{
		{"000000000000000-0-00-00000000000000", 0, 0, 0, 0, 0},
		{"000000000000a05-2-7f-00000000000000", 0xA, 5, 2, 0x7F, 0},
		{"0000000000fff0f-f-ff-00000000000000", 0xFFF, 0xF, 0xF, 0xFF, 0},
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

func BenchmarkStringComparison(b *testing.B) {
	uuid1 := New()
	uuid2 := New()
	s1 := uuid1.String()
	s2 := uuid2.String()
	for i := 0; i < b.N; i++ {
		_ = s1 == s2
	}
}

func BenchmarkByteArrayComparison(b *testing.B) {
	uuid1 := New()
	uuid2 := New()
	for i := 0; i < b.N; i++ {
		_ = uuid1 == uuid2
	}
}

func BenchmarkUUIDCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}
