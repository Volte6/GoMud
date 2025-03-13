package users

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// createTestIndexFile creates an index file with a fixed-width header and multiple user records.
// It uses the UserIndex receiver method formatFixedHeader so that the header is exactly 100 bytes long.
func createTestIndexFile(t *testing.T, filename string) *UserIndex {

	// Create an index instance for the test file.
	idx := &UserIndex{Filename: filename}
	idx.Create()

	// Prepare records: each iteration writes two records.
	for i := 1; i < 100; i++ {
		idx.AddUser(i*10, fmt.Sprintf(`alice_%d`, i))
		idx.AddUser(1000000+(i*10), fmt.Sprintf(`bob_%d`, i))
	}

	return idx
}

// TestSearchIndexByUsername tests the file-based search by creating a test index file
// and then verifying that lookups for "alice" and "bob" succeed while a lookup for a missing user ("charlie") fails.
func TestSearchIndexByUsername(t *testing.T) {

	// Create a temporary file for testing.
	tmpFile, err := os.CreateTemp("", "users_search_test_*.idx")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)
	//filename := "tmp.idx"

	// Write test index file data.
	idx := createTestIndexFile(t, filename)

	// Search for "alice_1" (should be found with userId 10).
	userId, found := idx.FindByUsername("alice_1")
	if !found || userId != 10 {
		t.Errorf("expected to find 'alice_1' with userId 10, got userId %d, found=%v", userId, found)
	}

	// Search for "alice_50" (should be found with userId 500).
	userId, found = idx.FindByUsername("alice_50")
	if !found || userId != 500 {
		t.Errorf("expected to find 'alice_50' with userId 500, got userId %d, found=%v", userId, found)
	}

	// Search for "bob_50" (should be found with userId 500).
	userId, found = idx.FindByUsername("bob_50")
	if !found || userId != 1000500 {
		t.Errorf("expected to find 'bob_50' with userId 500, got userId %d, found=%v", userId, found)
	}

	// Search for a non-existent user "charlie" (should not be found).
	userId, found = idx.FindByUsername("charlie")
	if found {
		t.Errorf("expected not to find 'charlie', but got userId %d", userId)
	}
}

// TestSearchIndexByUserId tests the file-based search by userId.
// It verifies that searching for an existing user id returns the correct username,
// that a non-existent user id returns ErrNotFound, and that appending a new record
// allows it to be found by its user id.
func TestSearchIndexByUserId(t *testing.T) {
	// Create a temporary file for testing.
	tmpFile, err := os.CreateTemp("", "users_search_by_userid_test_*.idx")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	// Write test index file data.
	idx := createTestIndexFile(t, filename)

	// Search for userId 10, which should correspond to "alice_1".
	username, found := idx.FindByUserId(10)
	if !found || username != "alice_1" {
		t.Errorf("expected to find 'alice_1' for userId 10, got username '%s', found=%v", username, found)
	}

	// Search for userId 500, which should correspond to "alice_50"
	// because the first record with userId 500 is "alice_50" (the "bob_50" record comes later).
	username, found = idx.FindByUserId(500)
	if !found || username != "alice_50" {
		t.Errorf("expected to find 'alice_50' for userId 500, got username '%s', found=%v", username, found)
	}

	// Search for a non-existent userId (e.g. 999999) and expect an error.
	username, found = idx.FindByUserId(999999)
	if found {
		t.Errorf("expected not to find userId 999999, but got username '%s'", username)
	}

	// Append a new record and then search for it by userId.
	if err := idx.AddUser(12345, "charlie"); err != nil {
		t.Fatalf("failed to append new record: %v", err)
	}
	username, found = idx.FindByUserId(12345)
	if !found || username != "charlie" {
		t.Errorf("expected to find 'charlie' for userId 12345, got username '%s', found=%v", username, found)
	}
}

// TestAppendUserRecord tests that appending a record updates the index file correctly:
// it verifies that the new record is searchable and that the header's record count is incremented.
func TestAppendUserRecord(t *testing.T) {

	// Create a temporary file for testing.
	tmpFile, err := os.CreateTemp("", "users_append_test_*.idx")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	// filename := "tmp.idx"

	// Write a test index file.
	idx := createTestIndexFile(t, filename)

	// Confirm that a user "charlie" does not exist yet.
	_, found := idx.FindByUsername("charlie")
	if found {
		t.Fatalf("expected 'charlie' not to be found before appending")
	}

	// Append the new record.
	if err := idx.AddUser(12345, "charlie"); err != nil {
		t.Fatalf("failed to append new record: %v", err)
	}

	// Search for "charlie" and verify it is found with the expected userId.
	userId, found := idx.FindByUsername("charlie")
	if !found {
		t.Fatalf("expected to find 'charlie' after appending")
	}
	if userId != 12345 {
		t.Errorf("expected userId 12345 for 'charlie', got %d", userId)
	}

	// Reopen the file and check the metadata header.
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open file for reading metadata: %v", err)
	}
	defer f.Close()

	// Use the index instance's readFixedHeader method.
	metaBytes, err := idx.readFixedHeader(f)
	if err != nil {
		t.Fatalf("failed to read metadata header: %v", err)
	}
	headerStr := strings.TrimSpace(string(metaBytes[:FixedHeaderTotalLength-1]))
	var version, recordCount, recordSize int
	if _, err := fmt.Sscanf(headerStr, "VERSION=%d,RECORDCOUNT=%d,RECORDSIZE=%d", &version, &recordCount, &recordSize); err != nil {
		t.Fatalf("failed to parse metadata header: %v", err)
	}

	// The createTestIndexFile function writes a header with a record count of 198.
	// After appending one record, we expect the record count to be 199.
	if recordCount != 199 {
		t.Errorf("expected record count to be 199 after appending, got %d", recordCount)
	}
}

// TestRemoveUserRecordByUsername tests that a record is removed correctly by updating the index file:
// it verifies that the removed record cannot be found, the header is updated, and that trying to remove
// a non-existent record returns ErrNotFound.
func TestRemoveUserRecordByUsername(t *testing.T) {
	// Create a temporary file for testing.
	tmpFile, err := os.CreateTemp("", "users_remove_test_*.idx")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	// Write test index file data.
	idx := createTestIndexFile(t, filename)

	// Remove an existing record: "alice_50".
	err = idx.RemoveByUsername("alice_50")
	if err != nil {
		t.Fatalf("failed to remove existing record 'alice_50': %v", err)
	}

	// Attempt to search for "alice_50" which should no longer exist.
	_, found := idx.FindByUsername("alice_50")
	if found {
		t.Errorf("expected 'alice_50' to be removed, but it was found")
	}

	// Open the file to check the updated metadata header.
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open file to read metadata: %v", err)
	}
	defer f.Close()

	metaBytes, err := idx.readFixedHeader(f)
	if err != nil {
		t.Fatalf("failed to read metadata header: %v", err)
	}
	headerStr := strings.TrimSpace(string(metaBytes[:FixedHeaderTotalLength-1]))
	var version, recordCount, recordSize int
	if _, err := fmt.Sscanf(headerStr, "VERSION=%d,RECORDCOUNT=%d,RECORDSIZE=%d", &version, &recordCount, &recordSize); err != nil {
		t.Fatalf("failed to parse metadata header: %v", err)
	}

	// createTestIndexFile creates a file with 198 records.
	// After removal, the record count should be 197.
	if recordCount != 197 {
		t.Errorf("expected record count to be 197 after removal, got %d", recordCount)
	}

	// Attempt to remove a non-existent record ("charlie") and expect ErrNotFound.
	err = idx.RemoveByUsername("charlie")
	if err == nil {
		t.Fatalf("expected error when removing non-existent record 'charlie'")
	}
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound when removing 'charlie', got: %v", err)
	}
}
