package dispatcher

import (
	"github.com/stretchr/testify/assert"
	"goa/lib/timex"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 自定义一个任务容器
type container struct {
	tasks    []int
	interval time.Duration
	execute  func(tasks interface{})
}

// 新建一个自定义容器
func newContainer(interval time.Duration, execute func(tasks interface{})) *container {
	return &container{
		interval: interval,
		execute:  execute,
	}
}

const threshold = 10

// Add 自定义新增方法
func (c *container) Add(task interface{}) bool {
	c.tasks = append(c.tasks, task.(int))
	return len(c.tasks) > threshold
}

// Exec 自定义执行方法
func (c *container) Execute(tasks interface{}) {
	if c.execute != nil {
		c.execute(tasks)
	} else {
		time.Sleep(c.interval)
	}
}

// PopAll 自定义弹出方法
func (c *container) PopAll() interface{} {
	tasks := c.tasks
	c.tasks = nil
	return tasks
}

func TestPeriodicalDispatcher_Sync(t *testing.T) {
	var done int32
	executor := NewPeriodicalDispatcher(time.Second, newContainer(500*time.Millisecond, nil))
	executor.Sync(func() {
		atomic.AddInt32(&done, 1)
	})
	assert.Equal(t, int32(1), atomic.LoadInt32(&done))
}

func TestPeriodicalDispatcher_QuitGoroutine(t *testing.T) {
	ticker := timex.NewFakeTicker()
	dispatcher := NewPeriodicalDispatcher(time.Millisecond, newContainer(time.Millisecond, nil))
	dispatcher.ticker = func() timex.Ticker {
		return ticker
	}

	// 当前存在的协程数量
	routines := runtime.NumGoroutine()

	dispatcher.Add(1)
	ticker.Tick()
	ticker.Wait(2 * idleRound * time.Millisecond)
	ticker.Tick()
	ticker.Wait(idleRound * time.Millisecond)

	assert.Equal(t, routines, runtime.NumGoroutine())
}

func TestPeriodicalDispatcher_Bulk(t *testing.T) {
	var vals []int
	var lock sync.Mutex // avoid data race

	dispatcher := NewPeriodicalDispatcher(time.Millisecond, newContainer(time.Millisecond, func(tasks interface{}) {
		t := tasks.([]int)
		for _, each := range t {
			lock.Lock()
			vals = append(vals, each)
			lock.Unlock()
		}
	}))

	ticker := timex.NewFakeTicker()
	dispatcher.ticker = func() timex.Ticker {
		return ticker
	}

	for i := 0; i < 10*threshold; i++ {
		if i%threshold == 5 {
			time.Sleep(2 * idleRound * time.Millisecond)
		}
		dispatcher.Add(i)
	}
	ticker.Tick()
	ticker.Wait(2 * idleRound * time.Millisecond)
	ticker.Tick()
	ticker.Tick()
	ticker.Wait(idleRound * time.Millisecond)

	var expect []int
	for i := 0; i < 10*threshold; i++ {
		expect = append(expect, i)
	}

	lock.Lock()
	assert.EqualValues(t, expect, vals)
	lock.Unlock()
}

func TestPeriodicalDispatcher_Wait(t *testing.T) {

}

func BenchmarkPeriodicalDispatcher_Add(b *testing.B) {
	b.ReportAllocs()

	executor := NewPeriodicalDispatcher(time.Second, newContainer(500*time.Millisecond, nil))
	for i := 0; i < b.N; i++ {
		executor.Add(i)
	}
}
