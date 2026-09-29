package timex

import "time"

const DefaultLayout = "2006-01-02 15:04:05"

// FormatUnix 将 unix 时间戳（秒或毫秒）格式化为 "2006-01-02 15:04:05" 字符串。
// 自动判断秒/毫秒：> 1e12 视为毫秒，否则视为秒。0 或负值返回空字符串。
func FormatUnix(ts int64) string {
	return FormatUnixLayout(ts, DefaultLayout)
}

// FormatUnixLayout 同 FormatUnix，但使用自定义 layout。
func FormatUnixLayout(ts int64, layout string) string {
	if ts <= 0 {
		return ""
	}
	if ts > 1e12 {
		return time.UnixMilli(ts).Format(layout)
	}
	return time.Unix(ts, 0).Format(layout)
}
