package dispatcher

import (
	"github.com/z-sdk/goa/lib/lang"
	"github.com/z-sdk/goa/lib/proc"
	"github.com/z-sdk/goa/lib/syncx"
	"github.com/z-sdk/goa/lib/threading"
	"github.com/z-sdk/goa/lib/timex"
	"reflect"
	"sync"
	"time"
)

const idleRound = 10

type (
	// TaskManager 任务管理者接口：负责任务的新增、执行、移除。
	TaskManager interface {
		Add(task interface{}) bool // 添加任务
		Execute(task interface{})  // 执行任务
		PopAll() interface{}       // 删除并返回当前所有任务
	}

	// Commander 指挥者：一个传递 interface{} 的通道
	Commander chan interface{}

	// Confirmer 确认者
	Confirmer chan lang.PlaceholderType

	// 定时调度器
	PeriodicalDispatcher struct {
		interval    time.Duration       // 任务调度间隔
		taskManager TaskManager         // 任务管理者
		commander   Commander           // 任务指挥者
		confirmer   Confirmer           // 任务确认者
		wg          sync.WaitGroup      // 同步等待组
		wgBarrier   syncx.Barrier       // 同步等待组的屏障器
		guarded     bool                // 是否守卫
		ticker      func() timex.Ticker // 任务断续器
		lock        sync.Mutex
	}
)

// NewPeriodicalDispatcher 定时执行器（间隔时间，任务管理器）
func NewPeriodicalDispatcher(interval time.Duration, taskManager TaskManager) *PeriodicalDispatcher {
	dispatcher := &PeriodicalDispatcher{
		interval:    interval,
		taskManager: taskManager,
		commander:   make(chan interface{}, 1),
		confirmer:   make(chan lang.PlaceholderType),
		ticker: func() timex.Ticker {
			return timex.NewTicker(interval)
		},
	}

	// 程序关闭前，尽量执行剩余任务
	proc.AddShutdownListener(func() {
		dispatcher.Flush()
	})

	return dispatcher
}

// Add 添加新任务给指挥者并确认可以执行
func (e *PeriodicalDispatcher) Add(task interface{}) {
	if tasks, ok := e.setAndGet(task); ok {
		e.commander <- tasks // 将当前所有任务发给指挥者
		<-e.confirmer        // 确认者进行确认
	}
}

// Flush 清洗任务
func (e *PeriodicalDispatcher) Flush() bool {
	e.enter()
	return e.execute(func() interface{} {
		e.lock.Lock()
		defer e.lock.Unlock()
		return e.taskManager.PopAll()
	}())
}

// Sync 同步执行一个自定义函数
func (e *PeriodicalDispatcher) Sync(fn func()) {
	e.lock.Lock()
	defer e.lock.Unlock()
	fn()
}

// Wait 加锁保护等待操作
func (e *PeriodicalDispatcher) Wait() {
	e.wgBarrier.Guard(func() {
		e.wg.Wait()
	})
}

// setAndGet 新增并返回任务，如有可能则后台直接执行任务
// 返回：加入后的所有待处理任务，是否已递交任务管理者
func (e *PeriodicalDispatcher) setAndGet(task interface{}) (interface{}, bool) {
	e.lock.Lock()
	defer func() {
		var start bool
		if !e.guarded {
			e.guarded = true
			start = true
		}
		e.lock.Unlock()
		if start {
			e.backgroundFlush()
		}
	}()

	if e.taskManager.Add(task) {
		return e.taskManager.PopAll(), true
	}

	return nil, false
}

// 后台任务清洗
func (e *PeriodicalDispatcher) backgroundFlush() {
	threading.GoSafe(func() {
		ticker := e.ticker()
		defer ticker.Stop()

		// 指挥者调度定时执行器
		var commanded bool
		lastTime := timex.Now()
		for {
			select {
			case tasks := <-e.commander:
				commanded = true
				e.enter()
				e.confirmer <- lang.Placeholder
				e.execute(tasks)
				lastTime = timex.Now()
			case <-ticker.Chan():
				if commanded {
					commanded = false
				} else if e.Flush() {
					lastTime = timex.Now()
				} else if timex.Since(lastTime) > e.interval*idleRound {
					e.lock.Lock()
					e.guarded = false
					e.lock.Unlock()

					// 再次清洗以防丢任务
					e.Flush()
					return
				}
			}
		}
	})
}

// enter 执行者进入，等待组加锁
func (e *PeriodicalDispatcher) enter() {
	e.wgBarrier.Guard(func() {
		e.wg.Add(1)
	})
}

// execute 调度任务管理者，执行任务
func (e *PeriodicalDispatcher) execute(tasks interface{}) bool {
	defer e.done()

	ok := e.has(tasks)
	if ok {
		e.taskManager.Execute(tasks)
	}

	return ok
}

// done 完成分组任务
func (e *PeriodicalDispatcher) done() {
	e.wg.Done()
}

// has 判断任务有无
func (e *PeriodicalDispatcher) has(tasks interface{}) bool {
	if tasks == nil {
		return false
	}

	val := reflect.ValueOf(tasks)
	switch val.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return val.Len() > 0
	default:
		// 其他类型默认为有值，让调用者自行处理
		return true
	}
}
