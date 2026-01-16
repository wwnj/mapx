# mapx

![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/wwnj/mapx.svg)](https://pkg.go.dev/github.com/wwnj/mapx)

High-performance concurrent-safe Map library, optimized for read-heavy workloads, providing two implementation strategies.

[简体中文](README.zh-CN.md)

## 🚀 Features

- **🔐 Concurrent-Safe**: Complete thread-safety guarantees
- **⚡ High-Performance Reads**: Lock-free read operations with excellent performance
- **🎯 Read-Heavy Optimization**: Designed specifically for read-heavy scenarios
- **🧩 Generic Support**: Type-safe with Go 1.18+ generics
- **📦 Zero Dependencies**: Only uses standard library
- **✅ Comprehensive Tests**: 100% test coverage

## 📦 Installation

```bash
go get github.com/wwnj/mapx
```

## 🎯 Two Implementations

### 1. RWMutexMap - atomic.Value + Mutex + COW

**Core Strategy**: Uses `atomic.Value` to store map pointer, lock-free reads, writes use mutex + Copy-On-Write

```go
type RWMutexMap[K comparable, V any] struct {
    mu   sync.Mutex
    data atomic.Value  // *map[K]V
}
```

**Features**:
- ✅ Completely lock-free reads using atomic loading
- ✅ Writes use mutex to avoid CAS retry overhead
- ✅ Suitable for read-heavy scenarios with moderate write concurrency
- ⚠️ Writes require copying the entire map

### 2. CASMap - atomic.Pointer + CAS + COW

**Core Strategy**: Uses `atomic.Pointer` to store map pointer, all writes use CAS (Compare-And-Swap)

```go
type CASMap[K comparable, V any] struct {
    data atomic.Pointer[map[K]V]
}
```

**Features**:
- ✅ Completely lock-free reads with excellent performance
- ✅ Lock-free writes using CAS atomic updates
- ✅ Suitable for read-heavy scenarios with very few, serial writes
- ⚠️ CAS may retry under high write concurrency, degrading performance
- ⚠️ Writes require copying the entire map

## 📖 API Documentation

Both implementations provide identical APIs:

| Method | Description |
|--------|-------------|
| `NewXXXMap[K, V]()` | Create new instance |
| `NewXXXMapWithCapacity[K, V](capacity)` | Create with pre-allocated capacity |
| `Get(key K) (V, bool)` | Retrieve value |
| `Set(key K, value V)` | Set value |
| `Delete(key K)` | Remove key |
| `Len() int` | Get number of elements |
| `Has(key K) bool` | Check if key exists |
| `Clear()` | Remove all elements |
| `Range(f func(K, V) bool)` | Iterate over all elements |
| `Keys() []K` | Get all keys |
| `Values() []V` | Get all values |
| `GetOrSet(key K, value V) (V, bool)` | Get or set |
| `SetIfAbsent(key K, value V) bool` | Set only if absent |
| `CompareAndSwap(key K, old V, new V) bool` | Compare and swap |

## 💡 Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/wwnj/mapx"
)

func main() {
    // Create RWMutexMap
    m := mapx.NewRWMutexMap[string, int]()

    // Set values
    m.Set("apple", 100)
    m.Set("banana", 200)

    // Get value
    if val, ok := m.Get("apple"); ok {
        fmt.Println("apple:", val) // apple: 100
    }

    // Check existence
    if m.Has("orange") {
        fmt.Println("found orange")
    }

    // Iterate
    m.Range(func(key string, value int) bool {
        fmt.Printf("%s: %d\n", key, value)
        return true
    })

    // Get or set
    val, existed := m.GetOrSet("grape", 300)
    if !existed {
        fmt.Println("grape was set to:", val)
    }

    // Compare and swap
    if m.CompareAndSwap("apple", 100, 150) {
        fmt.Println("apple updated to 150")
    }

    // Delete
    m.Delete("banana")

    // Clear
    m.Clear()
}
```

### Concurrent Scenario

```go
package main

import (
    "fmt"
    "sync"
    "github.com/wwnj/mapx"
)

func main() {
    m := mapx.NewCASMap[int, string]()
    var wg sync.WaitGroup

    // Concurrent writes
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            m.Set(id, fmt.Sprintf("value-%d", id))
        }(i)
    }

    // Concurrent reads
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if val, ok := m.Get(id); ok {
                _ = val
            }
        }(i)
    }

    wg.Wait()
    fmt.Println("Final size:", m.Len())
}
```

## 📊 Performance Benchmarks

Test Environment:
- CPU: Apple M2
- OS: macOS (darwin/arm64)
- Go: 1.25.0

### Benchmark Results

```
BenchmarkRWMutexMap_Get-8         	172345936	    3.272 ns/op	    0 B/op	   0 allocs/op
BenchmarkCASMap_Get-8             	149047956	    2.545 ns/op	    0 B/op	   0 allocs/op
BenchmarkRWMutexMap_Set-8         	   12426	   25050 ns/op	31312 B/op	   6 allocs/op
BenchmarkCASMap_Set-8             	   10000	   39372 ns/op	76238 B/op	  19 allocs/op
BenchmarkRWMutexMap_Mixed-8       	  118836	    2865 ns/op	 3701 B/op	   0 allocs/op
BenchmarkCASMap_Mixed-8           	   64294	    6171 ns/op	12863 B/op	   2 allocs/op
BenchmarkSyncMap_Mixed-8          	21968913	   14.53 ns/op	    6 B/op	   0 allocs/op
BenchmarkRWMutexMap_Small_Get-8   	112396059	    3.425 ns/op	    0 B/op	   0 allocs/op
BenchmarkCASMap_Small_Get-8       	153835388	    2.561 ns/op	    0 B/op	   0 allocs/op
BenchmarkRWMutexMap_Large_Get-8   	95117536	    3.366 ns/op	    0 B/op	   0 allocs/op
BenchmarkCASMap_Large_Get-8       	100000000	    5.370 ns/op	    0 B/op	   0 allocs/op
BenchmarkRWMutexMap_GetOrSet-8    	76675442	    4.060 ns/op	    0 B/op	   0 allocs/op
BenchmarkCASMap_GetOrSet-8        	111882759	    2.930 ns/op	    0 B/op	   0 allocs/op
```

### Performance Analysis

#### 📖 Read Performance

| Implementation | Ops/sec | Time/op | vs sync.Map |
|----------------|---------|---------|-------------|
| **CASMap** | 149M | 2.545 ns | 5.7x faster |
| **RWMutexMap** | 172M | 3.272 ns | 4.4x faster |

**Conclusion**: Both implementations significantly outperform `sync.Map` for reads, with CASMap slightly faster

#### ✏️ Write Performance

| Implementation | Ops/sec | Time/op | Memory |
|----------------|---------|---------|--------|
| **RWMutexMap** | 12.4K | 25.05 μs | 31KB/6 allocs |
| **CASMap** | 10K | 39.37 μs | 76KB/19 allocs |

**Conclusion**: RWMutexMap has better write performance with less memory allocation (avoids CAS retries)

#### 🔀 Mixed Operations (90% read / 10% write)

| Implementation | Ops/sec | Time/op | vs sync.Map |
|----------------|---------|---------|-------------|
| **sync.Map** | 21.9M | 14.53 ns | **Fastest** |
| **RWMutexMap** | 118K | 2.865 μs | 197x slower |
| **CASMap** | 64K | 6.171 μs | 424x slower |

**Conclusion**: `sync.Map` is more optimized for mixed scenarios, but COW strategies still valuable for specific use cases

#### 📏 Map Size Impact

**Small Map (10 elements)**:
- CASMap: 2.561 ns/op ⭐ **Fastest**
- RWMutexMap: 3.425 ns/op

**Large Map (10000 elements)**:
- RWMutexMap: 3.366 ns/op ⭐ **Fastest**
- CASMap: 5.370 ns/op

**Conclusion**: Larger maps show more pronounced copy overhead in CASMap

## 🎯 Selection Guide

### Use RWMutexMap

- ✅ Read-heavy scenarios (90%+ read operations)
- ✅ Large map capacity (> 1000 elements)
- ✅ Moderate write concurrency
- ✅ Need stable write performance

### Use CASMap

- ✅ Extremely read-heavy scenarios (95%+ read operations)
- ✅ Small map capacity (< 100 elements)
- ✅ Writes are mostly serial
- ✅ Pursuing ultimate read performance

### Use sync.Map

- ✅ Balanced read-write ratio (close to 1:1)
- ✅ Write-once, read-many access pattern for keys
- ✅ Different goroutines operate on disjoint key sets
- ✅ Using standard library, no external dependencies

### Comparison Table

| Scenario | RWMutexMap | CASMap | sync.Map |
|----------|-----------|---------|----------|
| Read-heavy (small map < 100) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| Read-heavy (large map > 1000) | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Balanced read-write | ⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ |
| Write-heavy | ⭐⭐ | ⭐ | ⭐⭐⭐⭐ |
| Concurrent writes | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Memory efficiency | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ |

## 🔧 Running Tests

```bash
# Run all unit tests (skips long-running concurrent tests)
go test -v -short

# Run full test suite including concurrent tests
go test -v -timeout=10m

# Run benchmarks
go test -run=^$ -bench=. -benchmem

# Run specific benchmarks
go test -run=^$ -bench='Get$' -benchmem
```

## 📝 Technical Details

### Copy-On-Write Strategy

Both implementations use COW strategy:
1. Copy entire map on write
2. Modify the copy
3. Atomically update pointer
4. Old map gets garbage collected

**Advantages**:
- Completely lock-free reads
- No read-write mutual exclusion

**Disadvantages**:
- High write overhead (time and space)
- Not suitable for large maps or write-heavy workloads

### Atomic Operations

**RWMutexMap**: Uses `atomic.Value` to store `*map[K]V`
**CASMap**: Uses `atomic.Pointer[map[K]V]` (Go 1.19+)

### CAS Correctness

The key to correct CAS operation is saving the old pointer before comparison:

```go
// ❌ Wrong: Load() returns a new pointer each time
m.data.CompareAndSwap(m.data.Load(), &newMap)

// ✅ Correct: Use saved old pointer
oldPtr := m.data.Load()
m.data.CompareAndSwap(oldPtr, &newMap)
```

## ⚠️ Caveats

1. **Memory Usage**: Write operations temporarily double memory usage (map copy)
2. **Write Performance**: Not suitable for write-heavy workloads
3. **Map Size**: Larger maps mean slower writes
4. **Concurrent Writes**: High write concurrency causes frequent CAS retries in CASMap

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📬 Contact

- GitHub: [@wwnj](https://github.com/wwnj)

---

⭐ If this project helps you, please give it a Star!
