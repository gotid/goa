package timex

import (
	"fmt"
	"time"
)

// MillisecondDuration 返回毫秒格式的时间段字符串
func MillisecondDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fms", float32(d)/float32(time.Millisecond))
}
