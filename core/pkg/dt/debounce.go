/**
 * @Author:         yi
 * @Description:    debounce-throttle
 * @Version:        1.0.0
 * @Date:           2025/2/11 17:00
 */
package dt

import (
	"time"

	cmap "github.com/orcaman/concurrent-map"
)

var debounceMaps = cmap.New()

// 定时任务模拟防抖函数
func Debounce(uniqueId string, interval time.Duration, call func()) {
	if uniqueId == "" || call == nil || interval <= 0 {
		return
	}

	var item = &debounceType{
		Call:     call,
		ExecTime: time.Now().UnixMilli() + interval.Milliseconds(),
	}
	if cache, ok := debounceMaps.Get(uniqueId); ok {
		if v, ok := cache.(*debounceType); ok {
			item.ExecTime = v.ExecTime
		}
	}

	debounceMaps.Set(uniqueId, item)
}

func debounceRunner() {
	for val := range time.NewTicker(time.Millisecond * 10).C {
		for uniqueId, item := range debounceMaps.Items() {
			var current = item.(*debounceType)
			if val.UnixMilli() >= current.ExecTime {
				go current.Call()
				debounceMaps.Remove(uniqueId)
			}
		}
	}
}
