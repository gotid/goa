package container

import (
	"fmt"
	"goa/lib/mathx"
	"goa/lib/timex"
	"strings"
	"sync"
)

const (
	numHistoryReasons = 5
	timeFormat        = "15:04:05"
)

// ErrorWindow 一扇观察最近5条记录的信息窗户
type ErrorWindow struct {
	reasons []string
	index   int
	count   int
	lock    sync.Mutex
}

func NewErrorWindow() *ErrorWindow {
	return &ErrorWindow{
		reasons: make([]string, numHistoryReasons),
	}
}

func (w *ErrorWindow) Add(reason string) {
	w.lock.Lock()
	w.reasons[w.index] = fmt.Sprintf("%s %s", timex.Time().Format(timeFormat), reason)
	w.index = (w.index + 1) % numHistoryReasons
	w.count = mathx.MinInt(w.count+1, numHistoryReasons)
	w.lock.Unlock()
}

// String 按时间降序输出
func (w *ErrorWindow) String() string {
	var reasons []string

	w.lock.Lock()
	// 反向排序
	for i := w.index - 1; i >= w.index-w.count; i-- {
		reasons = append(reasons, w.reasons[(i+numHistoryReasons)%numHistoryReasons])
	}
	w.lock.Unlock()

	return strings.Join(reasons, "\n")
}
