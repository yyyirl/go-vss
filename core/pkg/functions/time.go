package functions

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	TimeFormatYmd     = "ymd"
	TimeFormatYmdhis  = "ymdhis"
	TimeFormatYmdHIS  = "ymdHIS"
	TimeFormatISO8601 = "ISO8601"
	TimeFormatYm      = "ym"
)

type Timer struct{}

func NewTimer() *Timer {
	return &Timer{}
}

func (tm *Timer) Format(tp string) string {
	if tp == TimeFormatYmd {
		return time.Now().Format("2006-01-02")
	} else if tp == TimeFormatYmdhis {
		return time.Now().Format("2006-01-02-15-04-05")
	} else if tp == TimeFormatYmdHIS {
		return time.Now().Format("2006-01-0215:04:05")
	} else if tp == TimeFormatISO8601 {
		return time.Now().Format("20060102T15:04:05z")
	} else if tp == TimeFormatYm {
		return time.Now().Format("200601")
	}

	return time.Now().Format("2006-01-02 15:04:05")
}

func (tm *Timer) FormatTimestamp(timestamp int64, tp string) string {
	var d = time.Unix(timestamp, 0)
	if len(strconv.Itoa(int(timestamp))) > 11 {
		d = time.UnixMilli(timestamp)
	}

	if tp == TimeFormatYmd {
		return d.Format("2006-01-02")
	} else if tp == TimeFormatISO8601 {
		return d.Format("20060102T150405z")
	} else if tp == TimeFormatYmdhis {
		return d.Format("2006-01-02-15-04-05")
	} else if tp == TimeFormatYm {
		return d.Format("200601")
	}

	return d.Format("2006-01-02 15:04:05")
}

func (tm *Timer) FormatUTC() string {
	return time.Now().Format("2006-01-02[T]15:04:05[Z]")
}

func (tm *Timer) Day(tp string, num int) string {
	var t = time.Now().AddDate(0, 0, num)
	if tp == "ymd" {
		return t.Format("2006-01-02")
	}

	return t.Format("2006-01-02 15:04:05")
}

func (tm *Timer) DayWithTime(def time.Time, tp string, num int) string {
	var t = def.AddDate(0, 0, num)
	if tp == "ymd" {
		return t.Format("2006-01-02")
	}

	return t.Format("2006-01-02 15:04:05")
}

func (tm *Timer) DayTimestamp(num int) int64 {
	return time.Now().AddDate(0, 0, num).Unix()
}

func (tm *Timer) DayInitTimestamp(num int) int64 {
	var t = time.Now().AddDate(0, 0, num)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) MonthFirstDayTimestampMs() int64 {
	var t = time.Now()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).UnixMilli()
}

func (tm *Timer) MonthInitTimestamp(num int) int64 {
	var t = time.Now().AddDate(0, num, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) FewMonthTimestamp(it time.Time, num int) int64 {
	var t = it.AddDate(0, num, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location()).Unix()
}

func (tm *Timer) YearInitTimestamp(num int) int64 {
	var t = time.Now().AddDate(num, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) FewYearTimestamp(it time.Time, num int) int64 {
	var t = it.AddDate(num, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location()).Unix()
}

func (tm *Timer) HourInitTimestamp(num int) int64 {
	var t = time.Now().AddDate(0, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), num, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) MinuteInitTimestamp(num int) int64 {
	var t = time.Now().AddDate(0, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, num, 0, 0, t.Location()).Unix()
}

func (tm *Timer) WeekInitTimestamp(num int) int64 {
	return time.Now().Add(time.Duration(num) * 7 * 24 * time.Hour).Unix()
}

func (tm *Timer) FewWeekTimestamp(it time.Time, num int) int64 {
	return it.Add(time.Duration(num) * 7 * 24 * time.Hour).Unix()
}

func (tm *Timer) DayInitTimestampWithTime(initTime time.Time, num int) int64 {
	var t = initTime.AddDate(0, 0, num)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) MonthInitTimestampWithTime(initTime time.Time, num int) int64 {
	var t = initTime.AddDate(0, num, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) YearInitTimestampWithTime(initTime time.Time, num int) int64 {
	var t = initTime.AddDate(num, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) HourInitTimestampWithTime(initTime time.Time, num int) int64 {
	var t = initTime.AddDate(0, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), num, 0, 0, 0, t.Location()).Unix()
}

func (tm *Timer) MinuteInitTimestampWithTime(initTime time.Time, num int) int64 {
	var t = initTime.AddDate(0, 0, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, num, 0, 0, t.Location()).Unix()
}

func (tm *Timer) WeekInitTimestampWithTime(initTime time.Time, num int) int64 {
	return initTime.Add(time.Duration(num) * 7 * 24 * time.Hour).Unix()
}

func (tm *Timer) DateInitDayTime(date, tmp string) (time.Time, error) {
	var (
		t   time.Time
		err error
	)

	if tmp == "ymd" {
		t, err = time.ParseInLocation("2006-01-02", strings.Split(date, " ")[0], time.Local)
		if err != nil {
			return t, err
		}
	} else {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", strings.Split(date, " ")[0]+" 00:00:00", time.Local)
		if err != nil {
			return t, err
		}
	}

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()), nil
}

func (tm *Timer) DateInitDayTimestamp(date, tmp string) (int64, error) {
	d, err := tm.DateInitDayTime(date, tmp)
	if err != nil {
		return 0, err
	}

	return d.Unix(), nil
}

func (tm *Timer) DateInitWithTimestamp(timestamp int64) int64 {
	var t = time.Unix(timestamp, 0)
	year, month, day := t.Date()

	var (
		startOfDay          = time.Date(year, month, day, 0, 0, 0, 0, t.Location())
		startOfDayTimestamp = startOfDay.Unix()
	)

	return startOfDayTimestamp
}

func (tm *Timer) Now() int64 {
	return time.Now().Unix()
}

// 毫秒
func (tm *Timer) NowMilli() int64 {
	return time.Now().UnixMilli()
}

// 纳秒
func (tm *Timer) NowNano() int64 {
	return time.Now().UnixNano()
}

// 微秒
func (tm *Timer) UnixMicro() int64 {
	return time.Now().UnixMicro()
}

func (tm *Timer) FormatDate(tmp, date string) (int64, error) {
	var ts = "2006-01-02 15:04:05"
	if tmp == "ymd" {
		ts = "2006-01-02"
	}
	if tmp == "ymdT" { // 2025-08-12T07:26:45
		ts = "2006-01-02T15:04:05"
	}

	t, err := time.ParseInLocation(ts, date, time.Local)
	// t, err := time.Parse(ts, date)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

func (tm *Timer) FormatDateToMilli(tmp, date string) (int64, error) {
	var ts = "2006-01-02 15:04:05"
	if tmp == "ymd" {
		ts = "2006-01-02"
	}
	if tmp == "ymdT" { // 2025-08-12T07:26:45
		ts = "2006-01-02T15:04:05"
	}

	t, err := time.ParseInLocation(ts, date, time.Local)
	// t, err := time.Parse(ts, date)
	if err != nil {
		return 0, err
	}

	return t.UnixMilli(), nil
}

// 日期范围
func (tm *Timer) DateRange(num int, Type string) []string {
	if num <= 0 {
		return nil
	}

	var data = []string{
		tm.Day("ymd", 0),
	}
	for i := 1; i <= num; i++ {
		if Type == "forward" {
			data = append(data, tm.Day("ymd", i))
		} else {
			data = append(data, tm.Day("ymd", -i))
		}
	}

	return data
}

// 日期范围
func (tm *Timer) DateRangeTimestamp(num int) []int64 {
	if num <= 0 {
		return nil
	}

	var data []int64
	for i := 1; i <= num; i++ {
		data = append(data, tm.DayInitTimestamp(-i))
	}

	return data
}

// 指定日期 范围
func (tm *Timer) DateRangeFixed(date string, num int, Type string) ([]string, error) {
	t, err := tm.DateInitDayTime(date, "ymd")
	if err != nil {
		return nil, err
	}

	if num <= 0 {
		return nil, errors.New("num 不能为空")
	}

	var data []string
	for i := 0; i < num; i++ {
		if Type == "forward" {
			data = append(data, tm.DayWithTime(t, "ymd", i))
		} else {
			data = append(data, tm.DayWithTime(t, "ymd", -i))
		}
	}

	return data, nil
}

// 指定日期之间的天数
func (tm *Timer) DateRangeBetween(start, end, layout string) ([]string, error) {
	startDate, err := time.ParseInLocation(layout, start, time.Local)
	if err != nil {
		return nil, err
	}

	endDate, err := time.ParseInLocation(layout, end, time.Local)
	if err != nil {
		return nil, err
	}

	var (
		day  = int(endDate.Sub(startDate).Hours() / 24)
		data []string
	)
	data = append(data, start)

	for i := 1; i < day; i++ {
		data = append(data, startDate.AddDate(0, 0, i).Format(layout))
	}
	data = append(data, end)

	return data, nil
}

// 获取本周周一的日期
func (tm *Timer) GetFirstDateOfWeek(t time.Time) time.Time {
	offset := int(time.Monday - t.Weekday())
	if offset > 0 {
		offset = -6
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
}

// 获取下周周一
func (tm *Timer) GetNextFirstDateOfWeek(t time.Time) time.Time {
	return tm.GetFirstDateOfWeek(t).AddDate(0, 0, 7)
}

// 获取本周周日
func (tm *Timer) GetLastDateOfWeek(t time.Time) time.Time {
	return tm.GetFirstDateOfWeek(t).AddDate(0, 0, 6)
}

// 获取上周的周一日期
func (tm *Timer) GetLastWeekFirstDate(t time.Time) time.Time {
	return tm.GetFirstDateOfWeek(t).AddDate(0, 0, -7)
}

// 获取下周周日
func (tm *Timer) GetLastWeekLastDate(t time.Time) time.Time {
	return tm.GetLastDateOfWeek(t).AddDate(0, 0, 7)
}

// 当月最后一天
func (tm *Timer) GetNextWeekLastDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0).Add(-time.Second)
}

// 获取上个月1号
func (tm *Timer) GetLastMonthLastDate(t time.Time) time.Time {
	return tm.GetLastDateOfWeek(t).AddDate(0, -2, -1)
}

// 获取几个月前的时间
func (tm *Timer) GetFewMonthAgo(m uint) time.Time {
	return tm.GetLastDateOfWeek(time.Now()).AddDate(0, -int(m), -1)
}

// 指定时间月末
func (tm *Timer) GetLastMonthEndDate(t time.Time) time.Time {
	// 月初
	var startDate = t.Format("2006-01") + "-01"

	d, _ := time.ParseInLocation("2006-01-02", startDate, time.Local)
	return d.AddDate(0, 1, -1)
}

// 获取本月1号
func (tm *Timer) GetThisMonthLastDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	// return tm.GetLastDateOfWeek(t).AddDate(0, -1, -1)
}

// 获取本月1号
func (tm *Timer) GetFirstDayOfMonth(n int) time.Time {
	var date = time.Now().AddDate(0, n, 0)
	date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	return date
}

// 获取本月1号
func (tm *Timer) GetFirstDayOfFewMonth(n int) (list []time.Time) {
	for i := 0; i < n; i++ {
		var date = time.Now().AddDate(0, -i, 0)
		date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		list = append(list, date)
	}
	return
}

func (tm *Timer) RFC3339ToTimestamp(str string) int64 {
	t, _ := time.Parse(time.RFC3339, str)
	return t.Unix()
}

func (tm *Timer) InSameDay(timestamp1, timestamp2 int64) bool {
	time1 := time.Unix(timestamp1, 0)
	time2 := time.Unix(timestamp2, 0)

	// 判断是否处于同一天
	return time1.Year() == time2.Year() && time1.YearDay() == time2.YearDay()
}

// 几个月后的时间
func (tm *Timer) FewMonthLater(t time.Time, month int) int64 {
	return t.AddDate(0, month, 0).Unix()
}

// 几天后的时间
func (tm *Timer) FewDayLater(t time.Time, day int) int64 {
	return t.AddDate(0, 0, day).Unix()
}

// 几天后的时间
func (tm *Timer) FewDayInitLater(t time.Time, day int) time.Time {
	var d = t.AddDate(0, 0, day)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func (tm *Timer) TimestampWeekDay(timestamp int64) int {
	// 获取周几（0-6，周日为0）
	var d = time.Unix(timestamp, 0)
	if len(strconv.Itoa(int(timestamp))) > 11 {
		d = time.UnixMilli(timestamp)
	}

	return []int{7, 1, 2, 3, 4, 5, 6}[d.Weekday()]
}
