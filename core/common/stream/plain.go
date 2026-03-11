// @Title        plain
// @Description  main
// @Create       yiyiyi 2025/9/12 14:32

package stream

import (
	"fmt"
	"time"
)

type VideoPlain struct {
}

func NewVideoPlain() *VideoPlain {
	return &VideoPlain{}
}

func (v *VideoPlain) formatHour(hour int) string {
	return fmt.Sprintf("%02d:00", hour)
}

func (v *VideoPlain) Views(data [168]string) map[int][][2]string {
	result := make(map[int][][2]string)

	for day := 0; day < 7; day++ {
		dayIndex := day + 1
		dayData := data[day*24 : (day+1)*24]

		var timeRanges [][2]string
		start := -1

		for hour := 0; hour < 24; hour++ {
			if dayData[hour] == "1" {
				if start == -1 {
					start = hour
				}
			} else {
				if start != -1 {
					// 结束当前时间段：从start:00到(hour+1):00
					timeRanges = append(timeRanges, [2]string{
						v.formatHour(start),
						v.formatHour(hour),
					})
					start = -1
				}
			}
		}

		// 处理最后一个时间段（如果一直选中到23点）
		if start != -1 {
			timeRanges = append(timeRanges, [2]string{
				v.formatHour(start),
				"24:00",
			})
		}

		result[dayIndex] = timeRanges
	}

	return result
}

func (v *VideoPlain) GetTimeRanges(data [24]string, b int64) [][2]int64 {
	var t = time.Unix(b, 0)
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())

	var (
		result    [][2]int64
		startHour = -1
	)
	for hour := 0; hour < 24; hour++ {
		if data[hour] == "1" {
			if startHour == -1 {
				startHour = hour
			}
		} else {
			if startHour != -1 {
				startTime := startOfDay.Add(time.Duration(startHour) * time.Hour)
				endTime := startOfDay.Add(time.Duration(hour) * time.Hour)
				result = append(result, [2]int64{
					startTime.Unix(),
					endTime.Unix(),
				})
				startHour = -1
			}
		}
	}

	if startHour != -1 {
		startTime := startOfDay.Add(time.Duration(startHour) * time.Hour)
		endTime := startOfDay.Add(24 * time.Hour)
		result = append(result, [2]int64{
			startTime.Unix(),
			endTime.Unix(),
		})
	}

	return result
}
