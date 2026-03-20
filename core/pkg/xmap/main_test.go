// @Title        xmap_test
// @Description  线程安全测试
// @Create       yiyiyi 2026/3/21

package xmap

import (
	"sync"
	"testing"
)

func TestNew_GetSet_Remove_Len(t *testing.T) {
	var m = New[string, int](4)
	if m.Len() != 0 {
		t.Fatalf("期望 Len=0，实际 %d", m.Len())
	}

	m.Set("a", 1)
	var v, ok = m.Get("a")
	if !ok || v != 1 {
		t.Fatalf("期望 Get(a)=(1,true)，实际 (%v,%v)", v, ok)
	}

	m.Remove("a")
	if m.Contains("a") || m.Len() != 0 {
		t.Fatal("期望删除后不含 a 且 Len=0")
	}
}

func TestForEach_SnapshotAllowsReenterSet(t *testing.T) {
	var m = New[int, string](8)
	for i := range 5 {
		m.Set(i, "x")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		m.ForEach(func(key int, _ string) {
			if key == 2 {
				m.Set(100, "y")
			}
		})
	}()

	wg.Wait()

	var v, ok = m.Get(100)

	if !ok || v != "y" {
		t.Fatal("期望 ForEach 回调内可对同一 XMap 写入（快照遍历，不死锁）")
	}
}

func TestNilReceiver(t *testing.T) {
	var p *XMap[string, int]
	p.Set("a", 1)
	var _, ok = p.Get("a")
	if ok {
		t.Fatal("期望 nil 接收者 Get 为 false")
	}

	if p.Len() != 0 || p.Contains("") {
		t.Fatal("期望 nil Len=0、Contains 安全")
	}

	if p.All() != nil || p.Keys() != nil {
		t.Fatal("期望 nil 时 All/Keys 为 nil")
	}
}

func TestGetOrSet_SetIfAbsent(t *testing.T) {
	var (
		m  = New[string, int](0)
		v1 = m.GetOrSet("k", 10)
	)
	if v1 != 10 {
		t.Fatalf("期望首次 GetOrSet 返回 10，实际 %d", v1)
	}

	var v2 = m.GetOrSet("k", 20)
	if v2 != 10 {
		t.Fatalf("期望第二次 GetOrSet 仍返回已存在值 10，实际 %d", v2)
	}

	if !m.SetIfAbsent("k2", 7) || m.SetIfAbsent("k2", 8) {
		t.Fatal("期望 SetIfAbsent 首次 true、二次 false")
	}
}
