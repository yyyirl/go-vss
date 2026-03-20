/**
 * @Author:         yi
 * @Description:    线程安全泛型Map
 * @Version:        1.0.0
 * @Date:           2025/2/25 16:38
 */
package xmap

import "sync"

type (
	// XMap 泛型线程安全 Map；零值为 nil，须使用 New 创建后使用
	XMap[K comparable, V any] struct {
		mutex sync.RWMutex
		data  map[K]V
		// capHint 为构造时传入的容量提示
		capHint int
	}

	// RecordType 键值对记录，用于 Records / 快照遍历
	RecordType[K comparable, V any] struct {
		Key   K
		Value V
	}
)

// New 创建指定初始容量的XMap
func New[K comparable, V any](capacity int) *XMap[K, V] {
	var size = capacity
	if size < 0 {
		size = 0
	}

	return &XMap[K, V]{
		data:    make(map[K]V, size),
		capHint: size,
	}
}

// Set 设置键值对
func (x *XMap[K, V]) Set(key K, value V) {
	if x == nil {
		return
	}

	x.mutex.Lock()
	defer x.mutex.Unlock()

	x.data[key] = value
}

// Get 获取键对应的值及是否存在
func (x *XMap[K, V]) Get(key K) (V, bool) {
	var zero V
	if x == nil {
		return zero, false
	}

	x.mutex.RLock()
	defer x.mutex.RUnlock()

	value, ok := x.data[key]
	return value, ok
}

// Remove 删除指定键
func (x *XMap[K, V]) Remove(key K) {
	if x == nil {
		return
	}

	x.mutex.Lock()
	defer x.mutex.Unlock()

	delete(x.data, key)
}

// Clear 清空Map,按构造时的容量提示重新分配底层 map
func (x *XMap[K, V]) Clear() {
	if x == nil {
		return
	}

	x.mutex.Lock()
	defer x.mutex.Unlock()

	x.data = make(map[K]V, x.capHint)
}

// Keys 返回当前键的快照切片；顺序未定义空表返回 nil
func (x *XMap[K, V]) Keys() []K {
	if x == nil {
		return nil
	}

	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return nil
	}

	var keys = make([]K, 0, len(x.data))
	for k := range x.data {
		keys = append(keys, k)
	}

	return keys
}

// Values 返回所有值的切片
func (x *XMap[K, V]) Values() []V {
	if x == nil {
		return nil
	}

	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return nil
	}

	var values = make([]V, 0, len(x.data))
	for _, v := range x.data {
		values = append(values, v)
	}

	return values
}

// Records 返回键值对记录切片
func (x *XMap[K, V]) Records() []*RecordType[K, V] {
	if x == nil {
		return nil
	}

	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return nil
	}

	var records = make([]*RecordType[K, V], 0, len(x.data))

	for k, v := range x.data {
		records = append(records, &RecordType[K, V]{
			Key:   k,
			Value: v,
		})
	}

	return records
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
	if x == nil {
		return 0
	}

	x.mutex.RLock()
	defer x.mutex.RUnlock()

	return len(x.data)
}

// Contains 检查键是否存在
func (x *XMap[K, V]) Contains(key K) bool {
	if x == nil {
		return false
	}

	x.mutex.RLock()
	defer x.mutex.RUnlock()

	_, ok := x.data[key]
	return ok
}

// GetOrSet 若键存在则返回已有值
func (x *XMap[K, V]) GetOrSet(key K, defaultValue V) V {
	if x == nil {
		return defaultValue
	}

	x.mutex.Lock()
	defer x.mutex.Unlock()

	if val, ok := x.data[key]; ok {
		return val
	}

	x.data[key] = defaultValue
	return defaultValue
}

// SetIfAbsent 仅当键不存在时设置 value
func (x *XMap[K, V]) SetIfAbsent(key K, value V) bool {
	if x == nil {
		return false
	}

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
	if x == nil || fn == nil {
		return
	}

	var snap = x.snapshotRecords()
	for _, rec := range snap {
		fn(rec.Key, rec.Value)
	}
}

// snapshotRecords 在锁内复制键值到切片（值拷贝，非指针别名到新 RecordType）
func (x *XMap[K, V]) snapshotRecords() []RecordType[K, V] {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	if len(x.data) == 0 {
		return nil
	}

	var out = make([]RecordType[K, V], 0, len(x.data))
	for k, v := range x.data {
		out = append(out, RecordType[K, V]{
			Key:   k,
			Value: v,
		})
	}

	return out
}
