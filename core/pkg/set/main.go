// @Title        并发安全的泛型集合
// @Description  基于 map[T]struct{}
// @Create       yiyiyi 2025/9/5 09:03
package set

import "sync"

type (
	setMap[T comparable] map[T]struct{}

	CSet[T comparable] struct {
		data setMap[T]

		rw sync.RWMutex
	}
)

func New[T comparable](hint uint) *CSet[T] {
	return &CSet[T]{
		data: newSetMap[T](hint),
	}
}

// Add 将元素并入集合
func (m *CSet[T]) Add(elements ...T) {
	if m == nil || len(elements) == 0 {
		return
	}

	m.rw.Lock()
	defer m.rw.Unlock()

	m.data.add(elements...)
}

// Remove 从集合删除给定元素
func (m *CSet[T]) Remove(elements ...T) {
	if m == nil || len(elements) == 0 {
		return
	}

	m.rw.Lock()
	defer m.rw.Unlock()

	for _, ele := range elements {
		m.data.remove(ele)
	}
}

// Clear 清空集合
func (m *CSet[T]) Clear() {
	if m == nil {
		return
	}

	m.rw.Lock()
	defer m.rw.Unlock()

	m.data.clear()
}

// Contains 判断元素是否存在(读锁)
func (m *CSet[T]) Contains(ele T) bool {
	if m == nil {
		return false
	}

	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.data.contains(ele)
}

// IsEmpty 是否无任何元素
func (m *CSet[T]) IsEmpty() bool {
	if m == nil {
		return true
	}

	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.data.isEmpty()
}

// Size 返回元素个数
func (m *CSet[T]) Size() int {
	if m == nil {
		return 0
	}

	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.data.size()
}

// Range 对当前集合快照逐一枚举 f 返回 false 时提前结束
// 快照在持锁期间复制键,f 执行时已释放锁,因此 f 内可再次调用本 CSet 的其它方法(避免在持读锁的遍历回调里写同集合导致死锁)
func (m *CSet[T]) Range(f func(ele T) bool) {
	if m == nil || f == nil {
		return
	}

	var keys = m.snapshotKeys()
	for _, k := range keys {
		if !f(k) {
			break
		}
	}
}

// Values 返回当前集合元素切片(快照顺序未定义,与 map 遍历一致) nil 接收者返回 nil
func (m *CSet[T]) Values() []T {
	if m == nil {
		return nil
	}

	return m.snapshotKeys()
}

func (m *CSet[T]) snapshotKeys() []T {
	m.rw.RLock()
	defer m.rw.RUnlock()

	if len(m.data) == 0 {
		return nil
	}

	var out = make([]T, 0, len(m.data))
	for k := range m.data {
		out = append(out, k)
	}

	return out
}

// -------- 底层 setMap(无锁,仅由 CSet 在持锁下使用) --------

func newSetMap[T comparable](n uint) setMap[T] {
	return make(setMap[T], n)
}

func (s setMap[T]) add(elements ...T) {
	for _, item := range elements {
		s[item] = struct{}{}
	}
}

func (s setMap[T]) contains(ele T) bool {
	_, ok := s[ele]

	return ok
}

func (s setMap[T]) remove(ele T) {
	delete(s, ele)
}

func (s setMap[T]) isEmpty() bool {
	return len(s) == 0
}

func (s setMap[T]) size() int {
	return len(s)
}

func (s setMap[T]) clear() {
	for k := range s {
		delete(s, k)
	}
}
