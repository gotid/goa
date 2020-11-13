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

	// TaskChan 任务通道：一个传递 interface{} 的通道
	TaskChan chan interface{}

	// ConfirmChan 确认通道
	ConfirmChan chan lang.PlaceholderType

	// 定时调度器
	PeriodicalDispatcher struct {
		interval    time.Duration       // 任务调度间隔
		taskManager TaskManager         // 任务管理者
		taskChan    TaskChan            // 任务任务通道
		confirmChan ConfirmChan         // 任务确认通道
		wg          sync.WaitGroup      // 同步等待组
		wgBarrier   syncx.Barrier       // 同步等待组的屏障器
		guarded     bool                // 是否守卫
		ticker      func() timex.Ticker // 任务断续器
		lock        sync.Mutex
	}
)

// NewPeriodicalDispatcher 定时调度器（间隔时间，任务管理器）
func NewPeriodicalDispatcher(interval time.Duration, taskManager TaskManager) *PeriodicalDispatcher {
	dispatcher := &PeriodicalDispatcher{
		interval:    interval,
		taskManager: taskManager,
		taskChan:    make(chan interface{}, 1),
		confirmChan: make(chan lang.PlaceholderType),
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

// Add 添加新任务给任务通道并确认可以执行
func (pd *PeriodicalDispatcher) Add(task interface{}) {
	if tasks, ok := pd.setAndGet(task); ok {
		pd.taskChan <- tasks // 将当前所有任务发给任务通道
		<-pd.confirmChan     // 确认通道进行确认
	}
}

// Flush 清洗任务
func (pd *PeriodicalDispatcher) Flush() bool {
	pd.enter()
	return pd.execute(func() interface{} {
		pd.lock.Lock()
		defer pd.lock.Unlock()
		return pd.taskManager.PopAll()
	}())
}

// Sync 同步执行一个自定义函数
func (pd *PeriodicalDispatcher) Sync(fn func()) {
	pd.lock.Lock()
	defer pd.lock.Unlock()
	fn()
}

// Wait 加锁保护等待操作
func (pd *PeriodicalDispatcher) Wait() {
	pd.Flush()
	pd.wgBarrier.Guard(func() {
		pd.wg.Wait()
	})
}

// setAndGet 新增并返回任务，如有可能则后台直接执行任务
// 返回：加入后的所有待处理任务，是否已递交任务管理者
func (pd *PeriodicalDispatcher) setAndGet(task interface{}) (interface{}, bool) {
	pd.lock.Lock()
	defer func() {
		var start bool
		if !pd.guarded {
			pd.guarded = true
			start = true
		}
		pd.lock.Unlock()
		if start {
			pd.backgroundFlush()
		}
	}()

	if pd.taskManager.Add(task) {
		return pd.taskManager.PopAll(), true
	}

	return nil, false
}

// 后台任务清洗
func (pd *PeriodicalDispatcher) backgroundFlush() {
	threading.GoSafe(func() {
		ticker := pd.ticker()
		defer ticker.Stop()

		// 任务通道调度定时执行器
		var executed bool
		lastTime := timex.Now()
		for {
			select {
			case tasks := <-pd.taskChan:
				executed = true
				pd.enter()
				pd.confirmChan <- lang.Placeholder
				pd.execute(tasks)
				lastTime = timex.Now()
			case <-ticker.Chan():
				if executed {
					executed = false
				} else if pd.Flush() {
					lastTime = timex.Now()
				} else if timex.Since(lastTime) > pd.interval*idleRound {
					pd.lock.Lock()
					pd.guarded = false
					pd.lock.Unlock()

					// 再次清洗以防丢任务
					pd.Flush()
					return
				}
			}
		}
	})
}

// enter 执行者进入，等待组加锁
func (pd *PeriodicalDispatcher) enter() {
	pd.wgBarrier.Guard(func() {
		pd.wg.Add(1)
	})
}

// execute 调度任务管理者，执行任务
func (pd *PeriodicalDispatcher) execute(tasks interface{}) bool {
	defer pd.done()

	ok := pd.has(tasks)
	if ok {
		pd.taskManager.Execute(tasks)
	}

	return ok
}

// done 完成分组任务
func (pd *PeriodicalDispatcher) done() {
	pd.wg.Done()
}

// has 判断任务有无
func (pd *PeriodicalDispatcher) has(tasks interface{}) bool {
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
