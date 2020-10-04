package timex

import "time"

// TODO 没理解
// 使用足够长的过去时间作为起始时间，以防 time.Now() - lastTime 等于 0
var initTime = time.Now().AddDate(-1, -1, -1)

// Now 从初始时间至今的时间长度
func Now() time.Duration {
	return time.Since(initTime)
}

// Since 从 d 至今的时间差
func Since(d time.Duration) time.Duration {
	return time.Since(initTime) - d
}

// Time 当前时间
func Time() time.Time {
	return initTime.Add(Now())
}
