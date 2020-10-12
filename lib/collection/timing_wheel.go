package collection

import (
	"container/list"
	"fmt"
	"github.com/z-sdk/goa/lib/lang"
	"github.com/z-sdk/goa/lib/threading"
	"github.com/z-sdk/goa/lib/timex"
	"time"
)

const drainWorkers = 8

type (
	TimingWheel struct {
		interval      time.Duration // 时间划分刻度，一个最小的时间格子，是定时器的转动单位
		ticker        timex.Ticker
		slots         []*list.List
		timers        *SafeMap
		numSlots      int     // 时间轮的插槽数量（有几个圆）
		tickedPos     int     // 定时器位置
		execute       Execute // 时间点执行函数
		setChannel    chan timingEntry
		moveChannel   chan baseEntry
		removeChannel chan interface{}
		drainChannel  chan func(key, value interface{})
		stopChannel   chan lang.PlaceholderType
	}

	Execute func(key, value interface{})

	baseEntry struct {
		delay time.Duration
		key   interface{}
	}

	timingEntry struct {
		baseEntry
		value   interface{}
		circle  int
		diff    int
		removed bool
	}

	positionEntry struct {
		pos  int
		item *timingEntry
	}

	timingTask struct {
		key   interface{}
		value interface{}
	}
)

func NewTimingWheel(interval time.Duration, numSlots int, execute Execute) (*TimingWheel, error) {
	if interval <= 0 || numSlots <= 0 || execute == nil {
		return nil, fmt.Errorf("执行间隔(%v)/插槽数量(%d) 必须大于零，执行函数不能为空(%p)", interval, numSlots, execute)
	}

	return newTimingWheelWithClock(interval, numSlots, execute, timex.NewTicker(interval))
}

// 真正做初始化
func newTimingWheelWithClock(interval time.Duration, numSlots int, execute Execute, ticker timex.Ticker) (*TimingWheel, error) {
	w := &TimingWheel{
		interval:      interval,                                // 单个时间间隔
		ticker:        ticker,                                  // 定时器，做时间推动，以 interval 为单位推进
		slots:         make([]*list.List, numSlots),            // 时间槽，双向链表实现
		timers:        NewSafeMap(),                            // 存储 task{key, value} 的 map，提供 execute 执行函数所需参数
		numSlots:      numSlots,                                // 时间槽个数
		tickedPos:     numSlots - 1,                            // 位于上一次虚拟circle中
		execute:       execute,                                 // 时间点任务的真正执行函数
		setChannel:    make(chan timingEntry),                  // 设置任务的通道
		moveChannel:   make(chan baseEntry),                    // 移动任务的通道
		removeChannel: make(chan interface{}),                  // 移除任务的通道
		drainChannel:  make(chan func(key, value interface{})), // 执行任务通道
		stopChannel:   make(chan lang.PlaceholderType),         // 停止任务通道
	}

	w.initSlots()
	go w.run()

	return w, nil
}

// Drain 排水：向排水通道发送排水函数
func (w *TimingWheel) Drain(fn func(key, value interface{})) {
	w.drainChannel <- fn
}

func (w *TimingWheel) MoveTimer(key interface{}, delay time.Duration) {
	if delay <= 0 || key == nil {
		return
	}

	w.moveChannel <- baseEntry{
		delay: delay,
		key:   key,
	}
}

func (w *TimingWheel) RemoveTimer(key interface{}) {
	if key == nil {
		return
	}

	w.removeChannel <- key
}

func (w *TimingWheel) SetTimer(key, value interface{}, delay time.Duration) {
	if delay <= 0 || key == nil {
		return
	}

	w.setChannel <- timingEntry{
		baseEntry: baseEntry{
			delay: delay,
			key:   key,
		},
		value: value,
	}
}

func (w *TimingWheel) Stop() {
	close(w.stopChannel)
}

func (w *TimingWheel) initSlots() {
	for i := 0; i < w.numSlots; i++ {
		w.slots[i] = list.New()
	}
}

// 运行定时器，推动时间轮运转
func (w *TimingWheel) run() {
	for {
		select {
		// 定时器推动时间
		case <-w.ticker.Chan():
			w.onTick()
		//	收到新任务
		case task := <-w.setChannel:
			w.setTask(&task)
		//	移除任务
		case key := <-w.removeChannel:
			w.removeTask(key)
		//	清洗任务
		case fn := <-w.drainChannel:
			w.drainAll(fn)
		// 停止定时器
		case <-w.stopChannel:
			w.ticker.Stop()
			return
		}
	}
}

func (w *TimingWheel) onTick() {
	w.tickedPos = (w.tickedPos + 1) % w.numSlots
	l := w.slots[w.tickedPos]
	w.scanAndRunTasks(l)
}

func (w *TimingWheel) scanAndRunTasks(l *list.List) {
	var tasks []timingTask

	for e := l.Front(); e != nil; {
		task := e.Value.(*timingEntry)
		if task.removed {
			next := e.Next()
			l.Remove(e)
			w.timers.Del(task.key)
			e = next
			continue
		} else if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		} else if task.diff > 0 {
			next := e.Next()
			l.Remove(e)
			pos := (w.tickedPos + task.diff) % w.numSlots
			w.slots[pos].PushBack(task)
			w.setTimerPosition(pos, task)
			task.diff = 0
			e = next
			continue
		}

		tasks = append(tasks, timingTask{
			key:   task.key,
			value: task.value,
		})
		next := e.Next()
		l.Remove(e)
		w.timers.Del(task.key)
		e = next
	}

	w.runTasks(tasks)
}

func (w *TimingWheel) setTimerPosition(pos int, task *timingEntry) {
	if val, ok := w.timers.Get(task.key); ok {
		timer := val.(*positionEntry)
		timer.pos = pos
	} else {
		w.timers.Set(task.key, &positionEntry{
			pos:  pos,
			item: task,
		})
	}
}

func (w *TimingWheel) runTasks(tasks []timingTask) {
	if len(tasks) == 0 {
		return
	}

	go func() {
		for i := range tasks {
			threading.RunSafe(func() {
				w.execute(tasks[i].key, tasks[i].value)
			})
		}
	}()
}

func (w *TimingWheel) setTask(task *timingEntry) {
	if task.delay < w.interval {
		task.delay = w.interval
	}

	if val, ok := w.timers.Get(task.key); ok {
		entry := val.(*positionEntry)
		entry.item.value = task.value
		w.moveTask(task.baseEntry)
	} else {
		pos, circle := w.getPositionAndCircle(task.delay)
		task.circle = circle
		w.slots[pos].PushBack(task)
		w.setTimerPosition(pos, task)
	}
}

func (w *TimingWheel) moveTask(task baseEntry) {
	// 通过任务 key 名，获取 positionEntry 位置实体信息（位置信息、任务信息）
	val, ok := w.timers.Get(task.key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	// 任务的延迟时间比时间格间隔还要小，说明应该立即执行
	if task.delay < w.interval {
		threading.RunSafe(func() {
			w.execute(timer.item.key, timer.item.value)
		})
		return
	}

	pos, circle := w.getPositionAndCircle(task.delay)
	// 如果比时间格间隔大，则通过延迟时间算出时间轮中的新位置pos和circle
	if pos >= timer.pos {
		timer.item.circle = circle
		// 记录前后移动的偏移量，是为了后续重新入队
		timer.item.diff = pos - timer.pos
	} else if circle > 0 {
		// 移动到下一个内圆，将 circle 转换为 diff 的一部分
		circle--
		timer.item.circle = circle
		// 因为是一个数组，要加上 numSlots（相当于要走到下一层？）
		timer.item.diff = w.numSlots + pos - timer.pos
	} else {
		// 如果 offset 提前了，此时 task 也还在第一层
		// 标记删除老的 task，并重新入队，等待被执行
		timer.item.removed = true
		newTask := &timingEntry{
			baseEntry: task,
			value:     timer.item.value,
		}
		w.slots[pos].PushBack(newTask)
		w.setTimerPosition(pos, newTask)
	}
}

func (w *TimingWheel) getPositionAndCircle(delay time.Duration) (pos int, circle int) {
	steps := int(delay / w.interval)
	pos = (w.tickedPos + steps) % w.numSlots
	circle = (steps - 1) / w.numSlots

	return
}

func (w *TimingWheel) removeTask(key interface{}) {
	val, ok := w.timers.Get(key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	timer.item.removed = true
}

func (w *TimingWheel) drainAll(fn func(key interface{}, value interface{})) {
	runner := threading.NewTaskRunner(drainWorkers)
	for _, slot := range w.slots {
		for e := slot.Front(); e != nil; {
			task := e.Value.(*timingEntry)
			next := e.Next()
			slot.Remove(e)
			e = next
			if !task.removed {
				runner.Schedule(func() {
					fn(task.key, task.value)
				})
			}
		}
	}
}
