package dispatcher

import (
	"github.com/z-sdk/goa/lib/syncx"
	"github.com/z-sdk/goa/lib/timex"
	"time"
)

// ThrottleDispatcher 节流分发器
type ThrottleDispatcher struct {
	wait     time.Duration
	lastTime *syncx.AtomicDuration
}

// NewThrottleDispatcher 创建一个节流分发器，在 wait 秒内最多执行 func 一次。
func NewThrottleDispatcher(wait time.Duration) *ThrottleDispatcher {
	return &ThrottleDispatcher{
		wait:     wait,
		lastTime: syncx.NewAtomicDuration(),
	}
}

func (d *ThrottleDispatcher) DoOrDiscard(fn func()) bool {
	now := timex.Now()
	lastTime := d.lastTime.Load()
	if lastTime == 0 || lastTime+d.wait < now {
		d.lastTime.Set(now)
		fn()
		return true
	}
	return false
}
