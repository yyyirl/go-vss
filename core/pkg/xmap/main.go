/**
 * @Author:         yi
 * @Description:    main
 * @Version:        1.0.0
 * @Date:           2025/2/25 16:38
 */
package xmap

import (
	"sync"
)

type (
	XMap[K comparable, V any] struct {
		mutex sync.RWMutex
		data  map[K]V
		size  int
	}
	RecordType[K comparable, V any] struct {
		Key   K
		Value V
	}
)

func New[K comparable, V any](size int) *XMap[K, V] {
	return &XMap[K, V]{data: make(map[K]V, size), size: size}
}

func (x *XMap[K, V]) Set(key K, value V) {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	x.data[key] = value
}

func (x *XMap[K, V]) Get(key K) (V, bool) {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	value, ok := x.data[key]
	return value, ok
}

func (x *XMap[K, V]) Remove(key K) {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	delete(x.data, key)
}

func (x *XMap[K, V]) Clear() {
	x.mutex.Lock()
	defer x.mutex.Unlock()

	x.data = make(map[K]V, x.size)
}

func (x *XMap[K, V]) Keys() []K {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

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

func (x *XMap[K, V]) Values() []V {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

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

func (x *XMap[K, V]) Records() []*RecordType[K, V] {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

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

func (x *XMap[K, V]) All() map[K]V {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	return x.data
}

func (x *XMap[K, V]) Len() int {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	return len(x.data)
}

func (x *XMap[K, V]) Contains(key K) bool {
	x.mutex.RLock()
	defer x.mutex.RUnlock()

	_, ok := x.data[key]
	return ok
}
