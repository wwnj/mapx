# mapx

![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/wwnj/mapx.svg)](https://pkg.go.dev/github.com/wwnj/mapx)

高性能并发安全 Map 库，专为读多写少场景优化，提供两种实现策略。

[English](README.md) | 简体中文

## 🚀 特性

- **🔐 并发安全**: 完整的线程安全保证
- **⚡ 高性能读取**: 读操作完全无锁，性能极佳
- **🎯 读多写少优化**: 专为此类场景设计
- **🧩 泛型支持**: 基于 Go 1.18+ 泛型，类型安全
- **📦 零依赖**: 仅使用标准库
- **✅ 完整测试**: 100% 测试覆盖率

## 📦 安装

```bash
go get github.com/wwnj/mapx
```

## 🎯 两种实现

### 1. RWMutexMap - atomic.Value + Mutex + COW

**核心策略**: 使用 `atomic.Value` 存储 map 指针，读操作无锁，写操作使用互斥锁 + Copy-On-Write

```go
type RWMutexMap[K comparable, V any] struct {
    mu   sync.Mutex
    data atomic.Value  // *map[K]V
}
```

**特点**:
- ✅ 读操作完全无锁，使用原子加载
- ✅ 写操作使用互斥锁，避免 CAS 重试
- ✅ 适合读多写少，写操作有一定并发的场景
- ⚠️ 写时需要复制整个 map

### 2. CASMap - atomic.Pointer + CAS + COW

**核心策略**: 使用 `atomic.Pointer` 存储 map 指针，所有写操作使用 CAS（Compare-And-Swap）

```go
type CASMap[K comparable, V any] struct {
    data atomic.Pointer[map[K]V]
}
```

**特点**:
- ✅ 读操作完全无锁，性能极佳
- ✅ 写操作无锁，使用 CAS 原子更新
- ✅ 适合读频繁、写操作极少且串行的场景
- ⚠️ 高并发写入时 CAS 可能重试，性能下降
- ⚠️ 写时需要复制整个 map

## 📖 API 文档

两种实现提供完全一致的 API：

| 方法 | 说明 |
|------|------|
| `NewXXXMap[K, V]()` | 创建新实例 |
| `NewXXXMapWithCapacity[K, V](capacity)` | 创建并预分配容量 |
| `Get(key K) (V, bool)` | 获取 value |
| `Set(key K, value V)` | 设置 value |
| `Delete(key K)` | 删除 key |
| `Len() int` | 获取元素数量 |
| `Has(key K) bool` | 检查 key 是否存在 |
| `Clear()` | 清空所有元素 |
| `Range(f func(K, V) bool)` | 遍历所有元素 |
| `Keys() []K` | 获取所有 key |
| `Values() []V` | 获取所有 value |
| `GetOrSet(key K, value V) (V, bool)` | 获取或设置 |
| `SetIfAbsent(key K, value V) bool` | 仅在不存在时设置 |
| `CompareAndSwap(key K, old V, new V) bool` | 比较并交换 |

## 💡 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/wwnj/mapx"
)

func main() {
    // 创建 RWMutexMap
    m := mapx.NewRWMutexMap[string, int]()

    // 设置值
    m.Set("apple", 100)
    m.Set("banana", 200)

    // 获取值
    if val, ok := m.Get("apple"); ok {
        fmt.Println("apple:", val) // apple: 100
    }

    // 检查是否存在
    if m.Has("orange") {
        fmt.Println("found orange")
    }

    // 遍历
    m.Range(func(key string, value int) bool {
        fmt.Printf("%s: %d\n", key, value)
        return true
    })

    // 获取或设置
    val, existed := m.GetOrSet("grape", 300)
    if !existed {
        fmt.Println("grape was set to:", val)
    }

    // 比较并交换
    if m.CompareAndSwap("apple", 100, 150) {
        fmt.Println("apple updated to 150")
    }

    // 删除
    m.Delete("banana")

    // 清空
    m.Clear()
}
```

### 并发场景

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

    // 并发写入
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            m.Set(id, fmt.Sprintf("value-%d", id))
        }(i)
    }

    // 并发读取
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

## 📊 性能测试

测试环境：
- CPU: Apple M2
- OS: macOS (darwin/arm64)
- Go: 1.25.0

### Benchmark 结果

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

### 性能分析

#### 📖 读操作性能

| 实现 | 操作数/秒 | 每操作耗时 | 对比 sync.Map |
|------|-----------|-----------|--------------|
| **CASMap** | 149M | 2.545 ns | 5.7x 快 |
| **RWMutexMap** | 172M | 3.272 ns | 4.4x 快 |

**结论**: 两种实现的读性能都远超 `sync.Map`，CASMap 略快

#### ✏️ 写操作性能

| 实现 | 操作数/秒 | 每操作耗时 | 内存分配 |
|------|-----------|-----------|----------|
| **RWMutexMap** | 12.4K | 25.05 μs | 31KB/6次 |
| **CASMap** | 10K | 39.37 μs | 76KB/19次 |

**结论**: RWMutexMap 写性能更好，内存分配更少（避免 CAS 重试）

#### 🔀 混合操作性能 (90% 读 / 10% 写)

| 实现 | 操作数/秒 | 每操作耗时 | 对比 sync.Map |
|------|-----------|-----------|--------------|
| **sync.Map** | 21.9M | 14.53 ns | **最快** |
| **RWMutexMap** | 118K | 2.865 μs | 197x 慢 |
| **CASMap** | 64K | 6.171 μs | 424x 慢 |

**结论**: 混合场景下 `sync.Map` 性能更优，但 COW 策略在特定场景仍有价值

#### 📏 Map 大小影响

**小 Map (10 元素)**:
- CASMap: 2.561 ns/op ⭐ **最快**
- RWMutexMap: 3.425 ns/op

**大 Map (10000 元素)**:
- RWMutexMap: 3.366 ns/op ⭐ **最快**
- CASMap: 5.370 ns/op

**结论**: Map 越大，CASMap 的复制开销越明显

## 🎯 选型建议

### 使用 RWMutexMap

- ✅ 读多写少场景（90% 以上读操作）
- ✅ Map 容量较大（> 1000 元素）
- ✅ 写操作有一定并发
- ✅ 需要稳定的写性能

### 使用 CASMap

- ✅ 读极频繁，写极少（95% 以上读操作）
- ✅ Map 容量较小（< 100 元素）
- ✅ 写操作基本串行
- ✅ 追求极致读性能

### 使用 sync.Map

- ✅ 读写比例接近 1:1
- ✅ key 的写入模式是 "写一次，读多次"
- ✅ 不同 goroutine 操作不同的 key 集合
- ✅ 使用标准库，无需引入依赖

### 对比表

| 场景 | RWMutexMap | CASMap | sync.Map |
|------|-----------|---------|----------|
| 读多写少（小 map < 100） | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 读多写少（大 map > 1000） | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| 读写均衡 | ⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ |
| 写频繁 | ⭐⭐ | ⭐ | ⭐⭐⭐⭐ |
| 并发写 | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 内存效率 | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ |

## 🔧 运行测试

```bash
# 运行所有单元测试（跳过长时间并发测试）
go test -v -short

# 运行包含并发测试的完整测试
go test -v -timeout=10m

# 运行 benchmark
go test -run=^$ -bench=. -benchmem

# 运行特定 benchmark
go test -run=^$ -bench='Get$' -benchmem
```

## 📝 技术细节

### Copy-On-Write 策略

两种实现都采用 COW 策略：
1. 写操作时复制整个 map
2. 修改副本
3. 原子更新指针
4. 旧 map 由 GC 回收

**优点**:
- 读操作完全无锁
- 读写不互斥

**缺点**:
- 写操作开销大（时间和空间）
- 不适合大 map 或频繁写入

### 原子操作

**RWMutexMap**: 使用 `atomic.Value` 存储 `*map[K]V`
**CASMap**: 使用 `atomic.Pointer[map[K]V]` (Go 1.19+)

### CAS 正确性

CAS 操作的关键是保存旧指针并用它进行比较：

```go
// ❌ 错误：每次 Load() 返回新指针
m.data.CompareAndSwap(m.data.Load(), &newMap)

// ✅ 正确：使用保存的旧指针
oldPtr := m.data.Load()
m.data.CompareAndSwap(oldPtr, &newMap)
```

## ⚠️ 注意事项

1. **内存占用**: 写操作会临时增加一倍内存（复制 map）
2. **写性能**: 不适合写操作频繁的场景
3. **Map 大小**: Map 越大，写操作越慢
4. **并发写**: 高并发写入时，CASMap 可能频繁重试

## 🤝 Contributing

欢迎提交 Issue 和 Pull Request！

## 📬 联系

- GitHub: [@wwnj](https://github.com/wwnj)

---

⭐ 如果这个项目对你有帮助，请给个 Star！
