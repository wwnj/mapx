package mapx

import (
	"sync"
	"testing"
)

func TestRWMutexMap_BasicOperations(t *testing.T) {
	m := NewRWMutexMap[string, int]()

	// Test Set and Get
	m.Set("key1", 100)
	if val, ok := m.Get("key1"); !ok || val != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", val, ok)
	}

	// Test Get non-existent key
	if val, ok := m.Get("key2"); ok {
		t.Errorf("Expected (0, false), got (%d, true)", val)
	}

	// Test Has
	if !m.Has("key1") {
		t.Error("Expected key1 to exist")
	}
	if m.Has("key2") {
		t.Error("Expected key2 to not exist")
	}

	// Test Len
	if m.Len() != 1 {
		t.Errorf("Expected length 1, got %d", m.Len())
	}

	// Test Delete
	m.Delete("key1")
	if m.Has("key1") {
		t.Error("Expected key1 to be deleted")
	}
	if m.Len() != 0 {
		t.Errorf("Expected length 0, got %d", m.Len())
	}

	// Test Delete non-existent key (should not panic)
	m.Delete("nonexistent")
}

func TestRWMutexMap_GetOrSet(t *testing.T) {
	m := NewRWMutexMap[string, int]()

	// First call should set the value
	val, existed := m.GetOrSet("key1", 100)
	if existed || val != 100 {
		t.Errorf("Expected (100, false), got (%d, %v)", val, existed)
	}

	// Second call should return existing value
	val, existed = m.GetOrSet("key1", 200)
	if !existed || val != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", val, existed)
	}
}

func TestRWMutexMap_SetIfAbsent(t *testing.T) {
	m := NewRWMutexMap[string, int]()

	// Should set successfully
	if !m.SetIfAbsent("key1", 100) {
		t.Error("Expected SetIfAbsent to succeed")
	}

	// Should fail on second attempt
	if m.SetIfAbsent("key1", 200) {
		t.Error("Expected SetIfAbsent to fail")
	}

	// Value should remain unchanged
	if val, _ := m.Get("key1"); val != 100 {
		t.Errorf("Expected value 100, got %d", val)
	}
}

func TestRWMutexMap_CompareAndSwap(t *testing.T) {
	m := NewRWMutexMap[string, int]()

	// CAS on non-existent key should fail
	if m.CompareAndSwap("key1", 100, 200) {
		t.Error("Expected CAS to fail on non-existent key")
	}

	m.Set("key1", 100)

	// CAS with wrong old value should fail
	if m.CompareAndSwap("key1", 999, 200) {
		t.Error("Expected CAS to fail with wrong old value")
	}

	// CAS with correct old value should succeed
	if !m.CompareAndSwap("key1", 100, 200) {
		t.Error("Expected CAS to succeed")
	}

	// Verify new value
	if val, _ := m.Get("key1"); val != 200 {
		t.Errorf("Expected value 200, got %d", val)
	}
}

func TestRWMutexMap_Clear(t *testing.T) {
	m := NewRWMutexMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)
	m.Set("key3", 300)

	m.Clear()

	if m.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", m.Len())
	}
	if m.Has("key1") {
		t.Error("Expected all keys to be cleared")
	}
}

func TestRWMutexMap_Keys(t *testing.T) {
	m := NewRWMutexMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)
	m.Set("key3", 300)

	keys := m.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}
	if !keyMap["key1"] || !keyMap["key2"] || !keyMap["key3"] {
		t.Error("Keys not returned correctly")
	}
}

func TestRWMutexMap_Values(t *testing.T) {
	m := NewRWMutexMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)
	m.Set("key3", 300)

	values := m.Values()
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	valueMap := make(map[int]bool)
	for _, v := range values {
		valueMap[v] = true
	}
	if !valueMap[100] || !valueMap[200] || !valueMap[300] {
		t.Error("Values not returned correctly")
	}
}

func TestRWMutexMap_Range(t *testing.T) {
	m := NewRWMutexMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)
	m.Set("key3", 300)

	count := 0
	sum := 0
	m.Range(func(key string, value int) bool {
		count++
		sum += value
		return true
	})

	if count != 3 {
		t.Errorf("Expected to visit 3 entries, visited %d", count)
	}
	if sum != 600 {
		t.Errorf("Expected sum 600, got %d", sum)
	}

	// Test early termination
	count = 0
	m.Range(func(key string, value int) bool {
		count++
		return false // stop after first entry
	})
	if count != 1 {
		t.Errorf("Expected to visit 1 entry, visited %d", count)
	}
}

func TestRWMutexMap_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running concurrent test in short mode")
	}

	m := NewRWMutexMap[int, int]()
	const goroutines = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := id*iterations + j
				m.Set(key, key*2)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := id*iterations + j
				m.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify final count
	expectedLen := goroutines * iterations
	if m.Len() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, m.Len())
	}
}

func TestRWMutexMap_WithCapacity(t *testing.T) {
	m := NewRWMutexMapWithCapacity[string, int](100)
	m.Set("key1", 100)

	if val, ok := m.Get("key1"); !ok || val != 100 {
		t.Errorf("Expected (100, true), got (%d, %v)", val, ok)
	}
}
