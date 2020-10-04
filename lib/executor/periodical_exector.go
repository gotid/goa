package executor

import (
	"goa/lib/lang"
	"goa/lib/proc"
	"goa/lib/syncx"
	"goa/lib/threading"
	"goa/lib/timex"
	"reflect"
	"sync"
	"time"
)

const idleRound = 10

type (
	// Container 容器接口：定义任务的新增、执行、移除方法。
	Container interface {
		Add(task interface{}) bool // 添加任务到存储
		Exec(task interface{})     // 执行任务
		PopAll() interface{}       // 从容器删除并返回容器内所有任务
	}

	// Commander 指挥者：一个传递 interface{} 的通道
	Commander chan interface{}

	// Confirmer 确认者
	Confirmer chan lang.PlaceholderType

	// 定时任务执行器
	PeriodicalExecutor struct {
		container Container           // 任务存储者
		commander Commander           // 任务指挥者
		confirmer Confirmer           // 任务确认者
		interval  time.Duration       // 定期执行时间
		wg        sync.WaitGroup      // 同步等待组
		wgBarrier syncx.Barrier       // 同步等待组的屏障器
		guarded   bool                // 是否守卫
		newTicker func() timex.Ticker // 任务断续器
		lock      sync.Mutex
	}
)

// NewPeriodicalExecutor 定时执行器：定期调用执行容器内的任务
func NewPeriodicalExecutor(container Container, interval time.Duration) *PeriodicalExecutor {
	executor := &PeriodicalExecutor{
		container: container,
		commander: make(chan interface{}, 1),
		confirmer: make(chan lang.PlaceholderType),
		interval:  interval,
		newTicker: func() timex.Ticker {
			return timex.NewTicker(interval)
		},
	}

	// 程序关闭前，尽量执行完容器内的剩余任务
	proc.AddShutdownListener(func() {
		executor.Flush()
	})

	return executor
}

// Add 添加新任务给指挥者并确认可以执行
func (e *PeriodicalExecutor) Add(task interface{}) {
	if tasks, ok := e.setAndGet(task); ok {
		e.commander <- tasks // 将当前所有任务发给指挥者
		<-e.confirmer        // 确认者进行确认
	}
}

// Flush 清洗任务
func (e *PeriodicalExecutor) Flush() bool {
	e.enter()
	return e.execute(func() interface{} {
		e.lock.Lock()
		defer e.lock.Unlock()
		return e.container.PopAll()
	}())
}

// Sync 同步执行一个自定义函数
func (e *PeriodicalExecutor) Sync(fn func()) {
	e.lock.Lock()
	defer e.lock.Unlock()
	fn()
}

// Wait 加锁保护等待操作
func (e *PeriodicalExecutor) Wait() {
	e.wgBarrier.Guard(func() {
		e.wg.Wait()
	})
}

// setAndGet 新增并返回任务，如有可能则后台直接执行任务
// 返回：加入后的所有待处理任务，是否已递交容器
func (e *PeriodicalExecutor) setAndGet(task interface{}) (interface{}, bool) {
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

	if e.container.Add(task) {
		return e.container.PopAll(), true
	}

	return nil, false
}

// 后台任务清洗
func (e *PeriodicalExecutor) backgroundFlush() {
	threading.GoSafe(func() {
		ticker := e.newTicker()
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
func (e *PeriodicalExecutor) enter() {
	e.wgBarrier.Guard(func() {
		e.wg.Add(1)
	})
}

// execute 调用容器执行任务
func (e *PeriodicalExecutor) execute(tasks interface{}) bool {
	defer e.done()

	ok := e.has(tasks)
	if ok {
		e.container.Exec(tasks)
	}

	return ok
}

// done 完成分组任务
func (e *PeriodicalExecutor) done() {
	e.wg.Done()
}

// has 判断任务有无
func (e *PeriodicalExecutor) has(tasks interface{}) bool {
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
