package uuid

import (
	"strings"
	"testing"
)

func TestUUIDStringAndParse(t *testing.T) {
	gen := NewUUIDGenerator()
	typeID := uint8(42)
	uuid, err := gen.NewUUID(typeID)
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}
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
		"1234567890abcdef1-1-01",              // Only 3 parts.
		"1234567890abcdef1-1-01-xyz",          // Invalid unused part.
		"1234567890abcdef1-g-01-000000000000", // Invalid version part.
	}
	for _, s := range invalidInputs {
		if _, err := FromString(s); err == nil {
			t.Errorf("expected error for invalid UUID string %q, but got none", s)
		}
	}
}

func TestUniqueness(t *testing.T) {
	gen := NewUUIDGenerator()
	uuidSet := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		uuid, err := gen.NewUUID(1)
		if err != nil {
			t.Fatalf("failed to generate UUID: %v", err)
		}
		s := uuid.String()
		if _, exists := uuidSet[s]; exists {
			t.Errorf("duplicate UUID generated: %s", s)
		}
		uuidSet[s] = struct{}{}
	}
}

func TestUUIDStringFormat(t *testing.T) {
	gen := NewUUIDGenerator()
	uuid, err := gen.NewUUID(1)
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}
	// Expected string format: <timestamp><sequence>-<version>-<type>-<unused>
	// First part: 17 hex digits (16 for timestamp, 1 for sequence)
	// Second part: 1 hex digit (version)
	// Third part: 2 hex digits (type)
	// Fourth part: 12 hex digits (unused), must equal "000000000000"
	parts := strings.Split(uuid.String(), "-")
	if len(parts) != 4 {
		t.Fatalf("expected 4 parts, got %d", len(parts))
	}
	if len(parts[0]) != 17 {
		t.Errorf("expected first part to be 17 hex digits, got %d: %s", len(parts[0]), parts[0])
	}
	if len(parts[1]) != 1 {
		t.Errorf("expected version part to be 1 hex digit, got %d: %s", len(parts[1]), parts[1])
	}
	if len(parts[2]) != 2 {
		t.Errorf("expected type part to be 2 hex digits, got %d: %s", len(parts[2]), parts[2])
	}
	if parts[3] != unusedString {
		t.Errorf("expected unused part to be '%s', got %s", unusedString, parts[3])
	}
}

func TestFromStringUnusedNonZero(t *testing.T) {
	// Construct a string with a nonzero unused field.
	invalidStr := "00000000000000001-1-01-000000000001"
	if _, err := FromString(invalidStr); err == nil {
		t.Error("expected error for nonzero unused field, got nil")
	}
}

func TestUUIDExtractors(t *testing.T) {
	gen := NewUUIDGenerator()
	typeID := uint8(42)
	uuid, err := gen.NewUUID(typeID)
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}
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
	gen := NewUUIDGenerator()
	uuid, err := gen.NewUUID(1)
	if err != nil {
		t.Fatalf("failed to generate UUID: %v", err)
	}
	if uuid.IsNil() {
		t.Error("expected non-nil UUID to be IsNil() false")
	}
}
