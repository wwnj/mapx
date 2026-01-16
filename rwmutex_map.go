package mapx

import (
	"sync"
	"sync/atomic"
)

// RWMutexMap is a concurrent-safe Map implementation based on atomic.Value + Mutex + Copy-On-Write,
// optimized for read-heavy workloads.
//
// Read operations are completely lock-free, using atomic.Value to load the map pointer directly,
// providing excellent performance. Write operations use Mutex locking + Copy-On-Write strategy:
// acquire lock, copy map, modify, then atomically update.
//
// Advantages:
//   - Read operations are completely lock-free with excellent performance
//   - Write operations use locks instead of CAS, avoiding retry overhead
//   - Suitable for read-heavy scenarios, read performance better than traditional RWMutex
//
// Disadvantages:
//   - Write operations require copying the entire map, significant time and space overhead
//   - Not suitable for large maps or write-heavy scenarios
//
// Comparison with CASMap:
//   - Same read performance (both use lock-free atomic loading)
//   - Better write performance (lock guarantees mutual exclusion, no CAS retry overhead)
//   - Better suited for scenarios with moderate write concurrency but no retry desired
type RWMutexMap[K comparable, V any] struct {
	mu   sync.Mutex
	data atomic.Value // stores *map[K]V
}

// NewRWMutexMap creates a new RWMutexMap instance.
func NewRWMutexMap[K comparable, V any]() *RWMutexMap[K, V] {
	m := &RWMutexMap[K, V]{}
	newMap := make(map[K]V)
	m.data.Store(&newMap)
	return m
}

// NewRWMutexMapWithCapacity creates a new RWMutexMap instance with pre-allocated capacity.
// Pre-allocating capacity can reduce performance overhead from map growth.
func NewRWMutexMapWithCapacity[K comparable, V any](capacity int) *RWMutexMap[K, V] {
	m := &RWMutexMap[K, V]{}
	newMap := make(map[K]V, capacity)
	m.data.Store(&newMap)
	return m
}

// load atomically loads the current map pointer.
func (m *RWMutexMap[K, V]) load() map[K]V {
	return *m.data.Load().(*map[K]V)
}

// Get retrieves the value associated with the given key.
// Returns the zero value and false if the key doesn't exist; otherwise returns the value and true.
// Read operations are completely lock-free with excellent performance.
func (m *RWMutexMap[K, V]) Get(key K) (V, bool) {
	data := m.load()
	value, ok := data[key]
	return value, ok
}

// Set associates the given value with the given key.
// If the key already exists, the old value will be overwritten.
// Uses Mutex + Copy-On-Write strategy to avoid CAS retries.
func (m *RWMutexMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldMap := m.load()
	newMap := m.copyMap(oldMap)
	newMap[key] = value
	m.data.Store(&newMap)
}

// Delete removes the given key from the map.
// Has no effect if the key doesn't exist.
// Uses Mutex + Copy-On-Write strategy.
func (m *RWMutexMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldMap := m.load()
	// Return early if key doesn't exist to avoid unnecessary copy
	if _, ok := oldMap[key]; !ok {
		return
	}
	newMap := m.copyMap(oldMap)
	delete(newMap, key)
	m.data.Store(&newMap)
}

// Len returns the number of key-value pairs in the map.
func (m *RWMutexMap[K, V]) Len() int {
	data := m.load()
	return len(data)
}

// Has checks whether the given key exists in the map.
func (m *RWMutexMap[K, V]) Has(key K) bool {
	data := m.load()
	_, ok := data[key]
	return ok
}

// Clear removes all key-value pairs from the map.
func (m *RWMutexMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	newMap := make(map[K]V)
	m.data.Store(&newMap)
}

// Range iterates over all key-value pairs in the map.
// Calls f for each pair, stopping iteration if f returns false.
// Note: iteration is over a snapshot; concurrent writes don't affect the current iteration,
// so it's safe to call write methods within f without deadlock.
func (m *RWMutexMap[K, V]) Range(f func(key K, value V) bool) {
	data := m.load()
	for k, v := range data {
		if !f(k, v) {
			break
		}
	}
}

// Keys returns a slice containing all keys in the map.
func (m *RWMutexMap[K, V]) Keys() []K {
	data := m.load()
	keys := make([]K, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice containing all values in the map.
func (m *RWMutexMap[K, V]) Values() []V {
	data := m.load()
	values := make([]V, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

// GetOrSet retrieves the value for the given key, or sets it to the given value if it doesn't exist.
// Returns the value and true if the key already existed; otherwise returns the new value and false.
func (m *RWMutexMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	// Fast path: check if key exists without lock
	data := m.load()
	if v, ok := data[key]; ok {
		return v, true
	}

	// Key doesn't exist, acquire lock to set
	m.mu.Lock()
	defer m.mu.Unlock()
	oldMap := m.load()
	// Double-check to avoid race where another goroutine set the key while we waited for lock
	if v, ok := oldMap[key]; ok {
		return v, true
	}
	newMap := m.copyMap(oldMap)
	newMap[key] = value
	m.data.Store(&newMap)
	return value, false
}

// SetIfAbsent sets the value for the given key only if it doesn't already exist.
// Returns true if the value was set, false if the key already existed.
func (m *RWMutexMap[K, V]) SetIfAbsent(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldMap := m.load()
	if _, ok := oldMap[key]; ok {
		return false
	}
	newMap := m.copyMap(oldMap)
	newMap[key] = value
	m.data.Store(&newMap)
	return true
}

// CompareAndSwap atomically compares and swaps: sets newValue only if current value equals oldValue.
// Returns true if the swap succeeded, false if it failed (key doesn't exist or value doesn't match).
func (m *RWMutexMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldMap := m.load()
	v, ok := oldMap[key]
	if !ok || !compare(v, oldValue) {
		return false
	}
	newMap := m.copyMap(oldMap)
	newMap[key] = newValue
	m.data.Store(&newMap)
	return true
}

// copyMap creates a shallow copy of the map with all key-value pairs.
// This is the core implementation of the Copy-On-Write strategy.
func (m *RWMutexMap[K, V]) copyMap(oldMap map[K]V) map[K]V {
	newMap := make(map[K]V, len(oldMap))
	for k, v := range oldMap {
		newMap[k] = v
	}
	return newMap
}

// compare checks if two values are equal.
// Since generic types can't directly use == for non-comparable types,
// we use interface{} for comparison.
func compare[V any](a, b V) bool {
	return any(a) == any(b)
}
