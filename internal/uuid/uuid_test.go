package uuid

import (
	"fmt"
	"testing"
)

func TestUUIDStringAndParse(t *testing.T) {
	gen := newUUIDGenerator()
	typeID := IDType(42)
	u := gen.NewUUID(typeID)
	s := u.String()

	parsed, err := FromString(s)
	if err != nil {
		t.Fatalf("failed to parse UUID string: %v", err)
	}
	if u != parsed {
		t.Errorf("parsed UUID does not match original.\nGot:  %v\nWant: %v", parsed, u)
	}
}

func TestFromStringInvalidFormat(t *testing.T) {
	invalidInputs := []string{
		"1234",                                // Invalid length
		"1-234567890abcde-010000000000000000", // Only 3 parts
		"1-234567890abcde-01-xyz",             // 4 parts but invalid length
		"g-000000000000000-00-00000000000000", // Invalid version
	}
	for _, s := range invalidInputs {
		if _, err := FromString(s); err == nil {
			t.Errorf("expected error for invalid UUID string %q, but got none", s)
		}
	}
}

func TestUniqueness(t *testing.T) {
	gen := newUUIDGenerator()
	seen := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		u := gen.NewUUID(1)
		s := u.String()
		if _, exists := seen[s]; exists {
			t.Errorf("duplicate UUID generated: %s", s)
		}
		seen[s] = struct{}{}
	}
}

func TestUUIDExtractors(t *testing.T) {
	gen := newUUIDGenerator()
	typeID := IDType(42)
	u := gen.NewUUID(typeID)
	if u.Timestamp() == 0 {
		t.Error("Timestamp() returned 0; expected nonzero")
	}
	if u.Sequence() != 0 {
		t.Errorf("Sequence() = %d; expected 0", u.Sequence())
	}
	if u.Version() != currentVersion {
		t.Errorf("Version() = %d; expected %d", u.Version(), currentVersion)
	}
	if u.Type() != typeID {
		t.Errorf("Type() = %d; expected %d", u.Type(), typeID)
	}
	if u.Unused() != 0 {
		t.Errorf("Unused() = %d; expected 0", u.Unused())
	}
}

func TestIsNil(t *testing.T) {
	var nilUUID UUID
	if !nilUUID.IsNil() {
		t.Error("expected nil UUID to be IsNil() true")
	}
	if nilUUID.String() != `0` {
		t.Error("expected nil UUID String() to be 0")
	}
	gen := newUUIDGenerator()
	u := gen.NewUUID(1)
	if u.IsNil() {
		t.Error("expected non-nil UUID to be IsNil() false")
	}
}

func TestIsNilString(t *testing.T) {
	nilStr := "0anythingfollows"
	parsed, err := FromString(nilStr)
	if err != nil {
		t.Fatalf("failed to parse nil UUID string: %v", err)
	}
	var zero UUID
	if zero != parsed {
		t.Error("expected nil UUID parsed from string starting with 0 to equal nil UUID")
	}

	nilStr = ""
	parsed, err = FromString(nilStr)
	if err != nil {
		t.Fatalf("failed to parse nil UUID string: %v", err)
	}
	if zero != parsed {
		t.Error("expected nil UUID parsed from empty string to equal nil UUID")
	}
}

func TestFromStringRoundTrip(t *testing.T) {
	valid := []string{
		"1-000000000000000-00-00000000000002",
		"1-000000000000100-02-00000000000000",
		"1-000000000000100-02-01010101010101",
		fmt.Sprintf("%01x-%013x%02x-%02x-%014x", currentVersion, 0x2A, 0x03, 42, 0),
		fmt.Sprintf("%01x-%013x%02x-%02x-%014x", currentVersion, 0x10, 0xFF, 0xFF, 0),
	}
	for _, s := range valid {
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
		wantVer    uint8
		wantTS     uint64
		wantSeq    uint8
		wantType   IDType
		wantUnused uint64
	}
	fixtures := []fixture{
		{"0-000000000000000-00-00000000000000", 0, 0, 0, 0, 0},
		{"1-000000000000a05-7f-00000000000000", 1, 0xA, 0x05, 0x7F, 0},
		{"1-0000000000fff0f-ff-00000000000000", 1, 0xFFF, 0x0F, 0xFF, 0},
	}
	for _, f := range fixtures {
		u, err := FromString(f.str)
		if err != nil {
			t.Errorf("FromString(%q) error: %v", f.str, err)
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
	u1 := New()
	u2 := New()
	s1, s2 := u1.String(), u2.String()
	for i := 0; i < b.N; i++ {
		_ = s1 == s2
	}
}

func BenchmarkByteArrayComparison(b *testing.B) {
	u1 := New()
	u2 := New()
	for i := 0; i < b.N; i++ {
		_ = u1 == u2
	}
}

func BenchmarkUUIDCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}
