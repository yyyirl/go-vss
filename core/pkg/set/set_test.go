// @Title        set_test
// @Description  CSet 行为单测与基准
// @Create       yiyiyi 2026/3/20

package set

import (
	"sync"
	"testing"
)

func TestNew_AddContainsRemove_Size(t *testing.T) {
	var s = New[string](4)
	if s.Contains("a") {
		t.Fatal("期望：空集合不包含 a")
	}

	s.Add("a", "b")
	if s.Size() != 2 {
		t.Fatalf("期望 Size=2，实际 %d", s.Size())
	}

	if !s.Contains("a") || !s.Contains("b") {
		t.Fatal("期望：包含 a、b")
	}

	s.Remove("a")
	if s.Size() != 1 || s.Contains("a") || !s.Contains("b") {
		t.Fatal("期望：删除 a 后仅剩余 b")
	}

	s.Remove("x")
	if s.Size() != 1 {
		t.Fatal("期望：删除不存在元素为幂等")
	}

	s.Clear()
	if !s.IsEmpty() || s.Size() != 0 {
		t.Fatal("期望：Clear 后为空")
	}
}

func TestCSet_nil_Receivers(t *testing.T) {
	var (
		p *CSet[int]
	)

	if !p.IsEmpty() || p.Size() != 0 || p.Contains(0) {
		t.Fatal("期望：nil 接收者 Contains 为 false，IsEmpty 为 true，Size 为 0")
	}

	p.Add(1)
	p.Remove(1)
	p.Range(func(i int) bool {
		t.Fatal("期望：nil 接收者 Range 不执行回调")

		return true
	})

	if p.Values() != nil {
		t.Fatal("期望：nil 接收者 Values 返回 nil")
	}
}

func TestRange_Snapshot_AllowsReenterWrite(t *testing.T) {
	var s = New[int](8)
	for i := range 5 {
		s.Add(i)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		s.Range(func(ele int) bool {
			if ele == 2 {
				s.Add(100)
			}

			return true
		})
	}()

	wg.Wait()
	if !s.Contains(100) {
		t.Fatal("期望：Range 回调内可写同一集合（快照遍历，不死锁）")
	}
}

func TestValues_LengthMatchesSize(t *testing.T) {
	var s = New[int](0)
	s.Add(1, 2, 3)

	var v = s.Values()
	if len(v) != 3 {
		t.Fatalf("期望 Values 长度 3，实际 %d", len(v))
	}
}

func BenchmarkCSet_AddParallel(b *testing.B) {
	var s = New[int](1024)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int

		for pb.Next() {
			s.Add(i)

			i++
		}
	})
}

func BenchmarkCSet_ContainsParallel(b *testing.B) {
	var s = New[int](1024)
	for i := range 1024 {
		s.Add(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int

		for pb.Next() {
			_ = s.Contains(i % 1024)

			i++
		}
	})
}
