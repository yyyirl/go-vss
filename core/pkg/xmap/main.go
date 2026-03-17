/**
 * @Author:         yi
 * @Description:    线程安全泛型Map
 * @Version:        1.0.0
 * @Date:           2025/2/25 16:38
 */
package xmap

import "sync"

type (
	// XMap 泛型线程安全的Map结构
	XMap[K comparable, V any] struct {
		mutex sync.RWMutex // 读写锁，保证并发安全
		data  map[K]V      // 实际存储数据的map
		size  int          // 初始化容量，用于make优化
	}

	// RecordType 键值对记录类型
	RecordType[K comparable, V any] struct {
		Key   K
		Value V
	}
)

// New 创建指定初始容量的XMap
func New[K comparable, V any](size int) *XMap[K, V] {
	return &XMap[K, V]{
		data: make(map[K]V, size),
		size: size,
	}
}

// Set 设置键值对
func (x *XMap[K, V]) Set(key K, value V) {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	x.data[key] = value
}

// Get 获取键对应的值
func (x *XMap[K, V]) Get(key K) (V, bool) {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	value, ok := x.data[key]
	return value, ok
}

// Remove 删除指定键
func (x *XMap[K, V]) Remove(key K) {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	delete(x.data, key)
}

// Clear 清空Map
func (x *XMap[K, V]) Clear() {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	// 直接重新分配新map，避免逐个删除元素的开销
	x.data = make(map[K]V, x.size)
}

// Keys 返回所有键的切片
func (x *XMap[K, V]) Keys() []K {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return []K{}
	}

	var (
		keys = make([]K, len(x.data))
		i    = 0
	)
	for k := range x.data {
		keys[i] = k
		i++
	}

	return keys
}

// Values 返回所有值的切片
func (x *XMap[K, V]) Values() []V {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return []V{}
	}

	var (
		values = make([]V, len(x.data))
		i      = 0
	)
	for _, v := range x.data {
		values[i] = v
		i++
	}

	return values
}

// Records 返回所有键值对记录的切片
func (x *XMap[K, V]) Records() []*RecordType[K, V] {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return []*RecordType[K, V]{}
	}

	var (
		values = make([]*RecordType[K, V], len(x.data))
		i      = 0
	)
	for k, v := range x.data {
		values[i] = &RecordType[K, V]{
			Key:   k,
			Value: v,
		}
		i++
	}

	return values
}

// All 底层map
func (x *XMap[K, V]) All() map[K]V {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	return x.data
}

// AllCopy 返回底层map的副本
func (x *XMap[K, V]) AllCopy() map[K]V {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return nil
	}

	var result = make(map[K]V, len(x.data))
	for k, v := range x.data {
		result[k] = v
	}

	return result
}

// Len 返回元素个数
func (x *XMap[K, V]) Len() int {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	return len(x.data)
}

// Contains 检查键是否存在
func (x *XMap[K, V]) Contains(key K) bool {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	_, ok := x.data[key]
	return ok
}

// GetOrSet 获取值，如果不存在则设置默认值
func (x *XMap[K, V]) GetOrSet(key K, defaultValue V) V {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	if val, ok := x.data[key]; ok {
		return val
	}

	x.data[key] = defaultValue
	return defaultValue
}

// SetIfAbsent 如果键不存在则设置值
func (x *XMap[K, V]) SetIfAbsent(key K, value V) bool {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	if _, ok := x.data[key]; ok {
		return false
	}

	x.data[key] = value
	return true
}

// ForEach 遍历所有元素
func (x *XMap[K, V]) ForEach(fn func(key K, value V)) {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	for k, v := range x.data {
		fn(k, v)
	}
}
