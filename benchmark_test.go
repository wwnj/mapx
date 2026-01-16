package mapx

import (
	"sync"
	"testing"
)

// Benchmark for RWMutexMap - Read operations
func BenchmarkRWMutexMap_Get(b *testing.B) {
	m := NewRWMutexMap[int, int]()
	for i := 0; i < 1000; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 1000)
			i++
		}
	})
}

// Benchmark for CASMap - Read operations
func BenchmarkCASMap_Get(b *testing.B) {
	m := NewCASMap[int, int]()
	for i := 0; i < 1000; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 1000)
			i++
		}
	})
}

// Benchmark for sync.Map - Read operations
func BenchmarkSyncMap_Load(b *testing.B) {
	var m sync.Map
	for i := 0; i < 1000; i++ {
		m.Store(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Load(i % 1000)
			i++
		}
	})
}

// Benchmark for RWMutexMap - Write operations
func BenchmarkRWMutexMap_Set(b *testing.B) {
	m := NewRWMutexMap[int, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(i%1000, i)
			i++
		}
	})
}

// Benchmark for CASMap - Write operations
func BenchmarkCASMap_Set(b *testing.B) {
	m := NewCASMap[int, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(i%1000, i)
			i++
		}
	})
}

// Benchmark for sync.Map - Write operations
func BenchmarkSyncMap_Store(b *testing.B) {
	var m sync.Map

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i%1000, i)
			i++
		}
	})
}

// Benchmark for RWMutexMap - Mixed operations (90% read, 10% write)
func BenchmarkRWMutexMap_Mixed(b *testing.B) {
	m := NewRWMutexMap[int, int]()
	for i := 0; i < 1000; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				m.Set(i%1000, i)
			} else {
				m.Get(i % 1000)
			}
			i++
		}
	})
}

// Benchmark for CASMap - Mixed operations (90% read, 10% write)
func BenchmarkCASMap_Mixed(b *testing.B) {
	m := NewCASMap[int, int]()
	for i := 0; i < 1000; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				m.Set(i%1000, i)
			} else {
				m.Get(i % 1000)
			}
			i++
		}
	})
}

// Benchmark for sync.Map - Mixed operations (90% read, 10% write)
func BenchmarkSyncMap_Mixed(b *testing.B) {
	var m sync.Map
	for i := 0; i < 1000; i++ {
		m.Store(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				m.Store(i%1000, i)
			} else {
				m.Load(i % 1000)
			}
			i++
		}
	})
}

// Benchmark for RWMutexMap - Range operations
func BenchmarkRWMutexMap_Range(b *testing.B) {
	m := NewRWMutexMap[int, int]()
	for i := 0; i < 100; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Range(func(k, v int) bool {
			return true
		})
	}
}

// Benchmark for CASMap - Range operations
func BenchmarkCASMap_Range(b *testing.B) {
	m := NewCASMap[int, int]()
	for i := 0; i < 100; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Range(func(k, v int) bool {
			return true
		})
	}
}

// Benchmark for sync.Map - Range operations
func BenchmarkSyncMap_Range(b *testing.B) {
	var m sync.Map
	for i := 0; i < 100; i++ {
		m.Store(i, i*2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Range(func(k, v any) bool {
			return true
		})
	}
}

// Benchmark for RWMutexMap - Small map size (10 elements)
func BenchmarkRWMutexMap_Small_Get(b *testing.B) {
	m := NewRWMutexMap[int, int]()
	for i := 0; i < 10; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 10)
			i++
		}
	})
}

// Benchmark for CASMap - Small map size (10 elements)
func BenchmarkCASMap_Small_Get(b *testing.B) {
	m := NewCASMap[int, int]()
	for i := 0; i < 10; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 10)
			i++
		}
	})
}

// Benchmark for RWMutexMap - Large map size (10000 elements)
func BenchmarkRWMutexMap_Large_Get(b *testing.B) {
	m := NewRWMutexMap[int, int]()
	for i := 0; i < 10000; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 10000)
			i++
		}
	})
}

// Benchmark for CASMap - Large map size (10000 elements)
func BenchmarkCASMap_Large_Get(b *testing.B) {
	m := NewCASMap[int, int]()
	for i := 0; i < 10000; i++ {
		m.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 10000)
			i++
		}
	})
}

// Benchmark for RWMutexMap - GetOrSet
func BenchmarkRWMutexMap_GetOrSet(b *testing.B) {
	m := NewRWMutexMap[int, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.GetOrSet(i%1000, i)
			i++
		}
	})
}

// Benchmark for CASMap - GetOrSet
func BenchmarkCASMap_GetOrSet(b *testing.B) {
	m := NewCASMap[int, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.GetOrSet(i%1000, i)
			i++
		}
	})
}
