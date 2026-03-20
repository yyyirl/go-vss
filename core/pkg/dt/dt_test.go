// @Title        dt_test
// @Description  防抖/延迟/周期 行为校验
// @Create       yiyiyi 2026/3/20

package dt

import (
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// SetTimeout / SetInterval
// ---------------------------------------------------------------------------

func TestSetTimeout_CancelSkipsCallback(t *testing.T) {
	var fired int32

	var cancel = SetTimeout(100*time.Millisecond, func() {
		atomic.StoreInt32(&fired, 1)
	})

	cancel()

	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&fired) != 0 {
		t.Fatal("期望：取消 SetTimeout 后回调不执行")
	}
}

func TestSetInterval_FiresMultipleTimesBeforeCancel(t *testing.T) {
	var count int32

	var cancel = SetInterval(40*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(130 * time.Millisecond)

	cancel()

	time.Sleep(80 * time.Millisecond)

	var n = atomic.LoadInt32(&count)

	if n < 2 {
		t.Fatalf("期望：取消前时间窗内至少触发 2 次，实际 %d 次", n)
	}
}

// ---------------------------------------------------------------------------
// Debounce（全局 Ticker 扫描，执行后 Remove）
// ---------------------------------------------------------------------------

func TestDebounce_ResetsDeadlineOnRepeatCall(t *testing.T) {
	var count int32

	Debounce("db-test", 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(50 * time.Millisecond)

	Debounce("db-test", 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	// 第二次调用后仅过 60ms，未到「当前时刻 + interval」
	time.Sleep(60 * time.Millisecond)

	if atomic.LoadInt32(&count) != 0 {
		t.Fatalf("期望：推迟截止后此时仍未触发，实际 count=%d", count)
	}

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&count) != 1 {
		t.Fatalf("期望：静默结束后执行 1 次，实际 count=%d", count)
	}
}

// ---------------------------------------------------------------------------
// TrailingDebounce（每 key 独立 Timer，槽尾合并）
// ---------------------------------------------------------------------------

func TestTrailingDebounce_MergesToSingleExecution(t *testing.T) {
	var count int32

	TrailingDebounce("td-merge", 80*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	TrailingDebounce("td-merge", 80*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(40 * time.Millisecond)

	TrailingDebounce("td-merge", 80*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&count) != 1 {
		t.Fatalf("期望：多次触发仅 1 次执行，实际 %d 次", count)
	}
}

func TestTrailingDebounce_ReplacedScheduleDoesNotFireStale(t *testing.T) {
	var first, second int32

	TrailingDebounce("td-race", 100*time.Millisecond, func() {
		atomic.StoreInt32(&first, 1)
	})

	time.Sleep(30 * time.Millisecond)

	TrailingDebounce("td-race", 100*time.Millisecond, func() {
		atomic.StoreInt32(&second, 1)
	})

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&first) != 0 {
		t.Fatal("期望：已替换的调度不执行第一次回调")
	}

	if atomic.LoadInt32(&second) != 1 {
		t.Fatal("期望：仅最后一次调度在静默期结束后执行")
	}
}

// ---------------------------------------------------------------------------
// Throttle（前缘：窗口内仅首次）
// ---------------------------------------------------------------------------

func TestThrottle_LeadingEdgeOncePerWindow(t *testing.T) {
	var count int32

	Throttle("thr-window", 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	Throttle("thr-window", 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	Throttle("thr-window", 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&count) != 1 {
		t.Fatalf("期望：窗口内多次触发仅执行 1 次，实际 %d 次", count)
	}

	time.Sleep(100 * time.Millisecond)

	Throttle("thr-window", 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&count) != 2 {
		t.Fatalf("期望：窗口过期后可再触发，期望 count=2，实际 %d", count)
	}
}

// ---------------------------------------------------------------------------
// ThrottleFixedGridTrailing（固定栅格 + 槽尾执行 + 空闲 map 清理）
// ---------------------------------------------------------------------------

func TestThrottleFixedGridTrailing_SlotCountMatchesTimeline(t *testing.T) {
	var (
		period = 50 * time.Millisecond
		count  int32
		start  = time.Now()
	)

	for time.Since(start) < 102*time.Millisecond {
		ThrottleFixedGridTrailing("tfg-scale", period, func() {
			atomic.AddInt32(&count, 1)
		})
	}

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("期望：约 102ms / 50ms 栅格对齐共 3 次槽尾执行，实际 %d 次", count)
	}
}

func TestThrottleFixedGridTrailing_RemovesIdleKeyFromMap(t *testing.T) {
	var (
		key = "tfg-cleanup-key"
		n   int32
	)

	ThrottleFixedGridTrailing(key, 40*time.Millisecond, func() {
		atomic.AddInt32(&n, 1)
	})

	time.Sleep(120 * time.Millisecond)

	if throttleFixedGridMaps.Has(key) {
		t.Fatal("期望：槽尾执行完毕且无 pending/timer 后从 throttleFixedGridMaps 移除")
	}

	if atomic.LoadInt32(&n) != 1 {
		t.Fatalf("期望：回调执行 1 次，实际 %d 次", n)
	}
}

func TestThrottleFixedGridTrailing_LastCallWinsInSlot(t *testing.T) {
	var (
		period = 80 * time.Millisecond
		last   int32
	)

	for i := 0; i < 20; i++ {
		var v = int32(i)

		ThrottleFixedGridTrailing("tfg-last", period, func() {
			atomic.StoreInt32(&last, v)
		})
	}

	time.Sleep(period + 50*time.Millisecond)

	if atomic.LoadInt32(&last) != 19 {
		t.Fatalf("期望：同槽内最后一次闭包生效（19），实际 %d", last)
	}
}
