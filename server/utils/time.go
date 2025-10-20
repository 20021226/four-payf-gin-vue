package utils

import (
	"time"
)

func GetCurrentTime() int64 {
	return time.Now().Unix()
}

// FormatTime 自定义格式化时间函数
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func GetCurrentTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetTimeRangeStr 根据TTL计算格式化的开始时间和结束时间字符串
// layout: 时间格式，如 "20060102150405" 或 "2006-01-02 15:04:05"
// ttl: 时间范围，如果为0则默认5分钟
func GetTimeRange(ttl time.Duration) (startTime, endTime time.Time) {
	if ttl == 0 {
		ttl = time.Duration(5) * time.Minute
	}

	now := time.Now()
	startTime = now
	endTime = now.Add(ttl)

	return startTime, endTime
}

// GetTimeRangeStrWithExpires 根据expires秒数计算格式化的开始时间和结束时间字符串
// layout: 时间格式，如 "20060102150405" 或 "2006-01-02 15:04:05"
// expires: 过期时间（秒），如果为0则默认5分钟
func GetTimeRangeStrWithExpires(layout string, expires int) (startTimeStr, endTimeStr time.Time) {
	var ttl time.Duration
	if expires == 0 {
		ttl = time.Duration(5) * time.Minute
	} else {
		ttl = time.Duration(expires) * time.Second
	}

	return GetTimeRange(ttl)
}
