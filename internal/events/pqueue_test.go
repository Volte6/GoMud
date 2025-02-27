package events

import (
	"testing"
)

type MyItem struct {
	Name     string
	Priority int
}

// lessFunc defines items with a smaller Priority value as "higher priority."
func lessFunc(a, b MyItem) bool {
	return a.Priority < b.Priority
}

func TestEmptyQueue(t *testing.T) {
	queue := NewPriorityQueue(lessFunc)

	if queue.Len() != 0 {
		t.Errorf("Expected length 0, got %d", queue.Len())
	}

	// Test Peek on empty queue
	if item, ok := queue.Peek(); ok {
		t.Errorf("Expected no item, got %v", item)
	}

	// Test Poll on empty queue
	if item, ok := queue.Poll(); ok {
		t.Errorf("Expected no item, got %v", item)
	}
}

func TestSingleElementQueue(t *testing.T) {
	queue := NewPriorityQueue(lessFunc)

	item := MyItem{Name: "Single", Priority: 10}
	queue.Push(item)

	if queue.Len() != 1 {
		t.Errorf("Expected length 1, got %d", queue.Len())
	}

	// Test Peek
	peekedItem, ok := queue.Peek()
	if !ok {
		t.Error("Expected item after Peek, got none")
	} else if peekedItem != item {
		t.Errorf("Peek mismatch: expected %v, got %v", item, peekedItem)
	}

	// Test Poll
	polledItem, ok := queue.Poll()
	if !ok {
		t.Error("Expected item after Poll, got none")
	} else if polledItem != item {
		t.Errorf("Poll mismatch: expected %v, got %v", item, polledItem)
	}

	// Now queue should be empty
	if queue.Len() != 0 {
		t.Errorf("Expected length 0, got %d", queue.Len())
	}
}

func TestMultipleElements(t *testing.T) {
	// We expect items with lower Priority to come out first
	queue := NewPriorityQueue(lessFunc)

	items := []MyItem{
		{Name: "LowPriority", Priority: 100},
		{Name: "MediumPriority", Priority: 50},
		{Name: "HighPriority", Priority: 1},
		{Name: "HigherPriority", Priority: 0},
		{Name: "AnotherMedium", Priority: 50},
	}

	// Push items
	for _, it := range items {
		queue.Push(it)
	}

	// Queue length should match number of pushed items
	expectedLen := len(items)
	if queue.Len() != expectedLen {
		t.Errorf("Expected queue length %d, got %d", expectedLen, queue.Len())
	}

	// Poll items and check ordering
	// We expect the item with the smallest Priority first
	expectedOrder := []MyItem{
		{Name: "HigherPriority", Priority: 0},
		{Name: "HighPriority", Priority: 1},
		{Name: "MediumPriority", Priority: 50},
		{Name: "AnotherMedium", Priority: 50},
		{Name: "LowPriority", Priority: 100},
	}

	for i, expected := range expectedOrder {
		actual, ok := queue.Poll()
		if !ok {
			t.Fatalf("Expected item %v but queue was empty at index %d", expected, i)
		}
		if actual != expected {
			t.Errorf("Poll mismatch at index %d: expected %v, got %v", i, expected, actual)
		}
	}

	// After polling everything, queue should be empty
	if queue.Len() != 0 {
		t.Errorf("Expected queue length 0 after polling all items, got %d", queue.Len())
	}
}

func TestPeekDoesNotRemoveItem(t *testing.T) {
	queue := NewPriorityQueue(lessFunc)
	items := []MyItem{
		{Name: "Item1", Priority: 3},
		{Name: "Item2", Priority: 1},
	}

	for _, it := range items {
		queue.Push(it)
	}

	// Peek should return the item with the smallest Priority (Item2)
	peekedItem, ok := queue.Peek()
	if !ok {
		t.Error("Expected item after Peek, got none")
	}
	if peekedItem.Name != "Item2" || peekedItem.Priority != 1 {
		t.Errorf("Peek mismatch: expected Item2 with priority 1, got %v", peekedItem)
	}

	// Queue should still have 2 items after Peek
	if queue.Len() != 2 {
		t.Errorf("Expected queue length 2 after Peek, got %d", queue.Len())
	}
}

func TestOrderWithDuplicates(t *testing.T) {
	// If two items have the same priority, their insertion order
	// will determine how they appear in the linked-list-based approach.
	// But they should both appear in the correct position relative to
	// other items with different priorities.
	queue := NewPriorityQueue(lessFunc)

	duplicates := []MyItem{
		{Name: "Dup1", Priority: 5},
		{Name: "Dup2", Priority: 5},
		{Name: "Higher", Priority: 1},
		{Name: "Lower", Priority: 10},
	}

	for _, it := range duplicates {
		queue.Push(it)
	}

	// The first polled item should be "Higher" (priority=1)
	// Then "Dup1" (priority=5), then "Dup2" (priority=5), then "Lower" (priority=10)

	expectedOrder := []MyItem{
		{Name: "Higher", Priority: 1},
		{Name: "Dup1", Priority: 5},
		{Name: "Dup2", Priority: 5},
		{Name: "Lower", Priority: 10},
	}

	for i, expected := range expectedOrder {
		actual, ok := queue.Poll()
		if !ok {
			t.Fatalf("At index %d, expected %v but queue was empty", i, expected)
		}
		if actual != expected {
			t.Errorf("At index %d, expected %v, got %v", i, expected, actual)
		}
	}
	if queue.Len() != 0 {
		t.Errorf("Expected queue empty at end, got length %d", queue.Len())
	}
}
