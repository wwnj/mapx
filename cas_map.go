package mapx

import (
	"sync/atomic"
)

// CASMap is a concurrent-safe Map implementation based on CAS (Compare-And-Swap) + Copy-On-Write,
// particularly suitable for read-heavy workloads.
//
// Read operations are completely lock-free, using atomic operations to directly load the map pointer
// with excellent performance. Write operations use Copy-On-Write strategy: copy the entire map,
// modify it, then atomically update the pointer using CAS.
//
// Advantages:
//   - Read operations are completely lock-free with excellent performance
//   - No read-write mutual exclusion, reads never block
//   - Very suitable for read-heavy scenarios with frequent reads
//
// Disadvantages:
//   - Write operations require copying the entire map, significant time and space overhead
//   - Not suitable for large maps or write-heavy scenarios
//   - Under high write concurrency, CAS may fail and retry, degrading performance
type CASMap[K comparable, V any] struct {
	data atomic.Pointer[map[K]V]
}

// NewCASMap creates a new CASMap instance.
func NewCASMap[K comparable, V any]() *CASMap[K, V] {
	m := &CASMap[K, V]{}
	newMap := make(map[K]V)
	m.data.Store(&newMap)
	return m
}

// NewCASMapWithCapacity creates a new CASMap instance with pre-allocated capacity.
// Pre-allocating capacity can reduce performance overhead from map growth.
func NewCASMapWithCapacity[K comparable, V any](capacity int) *CASMap[K, V] {
	m := &CASMap[K, V]{}
	newMap := make(map[K]V, capacity)
	m.data.Store(&newMap)
	return m
}

// load atomically loads the current map pointer.
func (m *CASMap[K, V]) load() map[K]V {
	return *m.data.Load()
}

// Get retrieves the value associated with the given key.
// Returns the zero value and false if the key doesn't exist; otherwise returns the value and true.
// Read operations are completely lock-free with excellent performance.
func (m *CASMap[K, V]) Get(key K) (V, bool) {
	data := m.load()
	value, ok := data[key]
	return value, ok
}

// Set associates the given value with the given key.
// If the key already exists, the old value will be overwritten.
// Uses Copy-On-Write + CAS strategy with automatic retry on failure.
func (m *CASMap[K, V]) Set(key K, value V) {
	for {
		oldPtr := m.data.Load()
		oldMap := *oldPtr
		newMap := m.copyMap(oldMap)
		newMap[key] = value
		if m.data.CompareAndSwap(oldPtr, &newMap) {
			return
		}
		// CAS failed, retry
	}
}

// Delete removes the given key from the map.
// Has no effect if the key doesn't exist.
// Uses Copy-On-Write + CAS strategy with automatic retry on failure.
func (m *CASMap[K, V]) Delete(key K) {
	for {
		oldPtr := m.data.Load()
		oldMap := *oldPtr
		// Return early if key doesn't exist
		if _, ok := oldMap[key]; !ok {
			return
		}
		newMap := m.copyMap(oldMap)
		delete(newMap, key)
		if m.data.CompareAndSwap(oldPtr, &newMap) {
			return
		}
		// CAS failed, retry
	}
}

// Len returns the number of key-value pairs in the map.
func (m *CASMap[K, V]) Len() int {
	data := m.load()
	return len(data)
}

// Has checks whether the given key exists in the map.
func (m *CASMap[K, V]) Has(key K) bool {
	data := m.load()
	_, ok := data[key]
	return ok
}

// Clear removes all key-value pairs from the map.
func (m *CASMap[K, V]) Clear() {
	newMap := make(map[K]V)
	m.data.Store(&newMap)
}

// Range iterates over all key-value pairs in the map.
// Calls f for each pair, stopping iteration if f returns false.
// Note: iteration is over a snapshot; concurrent writes don't affect the current iteration.
func (m *CASMap[K, V]) Range(f func(key K, value V) bool) {
	data := m.load()
	for k, v := range data {
		if !f(k, v) {
			break
		}
	}
}

// Keys returns a slice containing all keys in the map.
func (m *CASMap[K, V]) Keys() []K {
	data := m.load()
	keys := make([]K, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice containing all values in the map.
func (m *CASMap[K, V]) Values() []V {
	data := m.load()
	values := make([]V, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

// GetOrSet retrieves the value for the given key, or sets it to the given value if it doesn't exist.
// Returns the value and true if the key already existed; otherwise returns the new value and false.
func (m *CASMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	// Fast path: check if key exists
	data := m.load()
	if v, ok := data[key]; ok {
		return v, true
	}

	// Key doesn't exist, use CAS to set
	for {
		oldPtr := m.data.Load()
		oldMap := *oldPtr
		// Double-check
		if v, ok := oldMap[key]; ok {
			return v, true
		}
		newMap := m.copyMap(oldMap)
		newMap[key] = value
		if m.data.CompareAndSwap(oldPtr, &newMap) {
			return value, false
		}
		// CAS failed, retry
	}
}

// SetIfAbsent sets the value for the given key only if it doesn't already exist.
// Returns true if the value was set, false if the key already existed.
func (m *CASMap[K, V]) SetIfAbsent(key K, value V) bool {
	for {
		oldPtr := m.data.Load()
		oldMap := *oldPtr
		if _, ok := oldMap[key]; ok {
			return false
		}
		newMap := m.copyMap(oldMap)
		newMap[key] = value
		if m.data.CompareAndSwap(oldPtr, &newMap) {
			return true
		}
		// CAS failed, retry
	}
}

// CompareAndSwap atomically compares and swaps: sets newValue only if current value equals oldValue.
// Returns true if the swap succeeded, false if it failed (key doesn't exist or value doesn't match).
func (m *CASMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) bool {
	for {
		oldPtr := m.data.Load()
		oldMap := *oldPtr
		v, ok := oldMap[key]
		if !ok || !compare(v, oldValue) {
			return false
		}
		newMap := m.copyMap(oldMap)
		newMap[key] = newValue
		if m.data.CompareAndSwap(oldPtr, &newMap) {
			return true
		}
		// CAS failed, retry
	}
}

// copyMap creates a shallow copy of the map with all key-value pairs.
// This is the core implementation of the Copy-On-Write strategy.
func (m *CASMap[K, V]) copyMap(oldMap map[K]V) map[K]V {
	newMap := make(map[K]V, len(oldMap))
	for k, v := range oldMap {
		newMap[k] = v
	}
	return newMap
}
