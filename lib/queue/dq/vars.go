package dq

import "time"

const (
	PriorityHigh   = 1 // 高优先级
	PriorityNormal = 2 // 中优先级
	PriorityLow    = 3 // 低优先级

	defaultTimeToRun = 5 * time.Second // 剩余运行时间 TTR
	reverseTimeout   = 5 * time.Second // 反向取值时间

	idSep   = "," // 编号分隔符
	timeSep = '/' // 时间分隔符
)
