/**
 * @Author:         yi
 * @Description:    throttled
 * @Version:        1.0.0
 * @Date:           2025/2/11 17:11
 */

package dt

import (
	"time"

	cmap "github.com/orcaman/concurrent-map"
)

var throttledMaps = cmap.New()

// 节流函数 一段时间内调用多次 只执行最后一次
func Throttled(uniqueId string, duration time.Duration, call func()) {
	if uniqueId == "" || call == nil || duration <= 0 {
		return
	}

	// 取消任务
	if val, ok := throttledMaps.Get(uniqueId); ok {
		if item, ok := val.(*throttledType); ok {
			item.Cancel()
		}
	}

	// 创建任务
	throttledMaps.Set(uniqueId, &throttledType{
		Call: call,
		Cancel: SetTimeout(duration, func() {
			if val, ok := throttledMaps.Get(uniqueId); ok {
				if item, ok := val.(*throttledType); ok {
					go item.Call()
					throttledMaps.Remove(uniqueId)
				}
			}
		}),
	})
}
