package set

import (
	"iter"
	"sync"
)

type (
	Set[T comparable] map[T]struct{}

	CSet[T comparable] struct {
		data Set[T]
		rw   sync.RWMutex
	}
)

// --- MARK CSet methods

func New[T comparable](n uint) *CSet[T] {
	return &CSet[T]{
		data: newSet[T](n),
	}
}

func (m *CSet[T]) Add(elements ...T) {
	m.rw.Lock()
	defer m.rw.Unlock()
	m.data.add(elements...)
}

func (m *CSet[T]) Remove(ele T) {
	m.rw.Lock()
	defer m.rw.Unlock()
	m.data.remove(ele)
}

func (m *CSet[T]) clear() {
	m.rw.Lock()
	defer m.rw.Unlock()
	m.data.clear()
}

func (m *CSet[T]) Contains(ele T) bool {
	m.rw.RLock()
	defer m.rw.RUnlock()
	return m.data.contains(ele)
}

func (m *CSet[T]) IsEmpty() bool {
	m.rw.RLock()
	defer m.rw.RUnlock()
	return m.data.isEmpty()
}

func (m *CSet[T]) Size() int {
	m.rw.RLock()
	defer m.rw.RUnlock()
	return m.data.size()
}

func (m *CSet[T]) Range(f func(ele T) bool) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	for k := range m.data {
		if !f(k) {
			break
		}
	}
}

func (m *CSet[T]) Values() []T {
	m.rw.RLock()
	defer m.rw.RUnlock()

	var (
		length = m.data.size()
		values = make([]T, length, length)

		i = 0
	)
	for k := range m.data.values() {
		values[i] = k
		i++
	}

	return values
}

// --- MARK set methods

func newSet[T comparable](n uint) Set[T] {
	return make(Set[T], n)
}

func (set Set[T]) add(elements ...T) {
	for _, item := range elements {
		set[item] = struct{}{}
	}
}

func (set Set[T]) contains(ele T) bool {
	_, ok := set[ele]
	return ok
}

func (set Set[T]) remove(ele T) {
	delete(set, ele)
}

func (set Set[T]) isEmpty() bool {
	return len(set) == 0
}

func (set Set[T]) size() int {
	return len(set)
}

func (set Set[T]) clear() {
	for k := range set {
		delete(set, k)
	}
}

func (set Set[T]) values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range set {
			if !yield(v) {
				return
			}
		}
	}
}
