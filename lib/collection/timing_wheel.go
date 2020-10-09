package collection

import (
	"container/list"
	"fmt"
	"goa/lib/lang"
	"goa/lib/threading"
	"goa/lib/timex"
	"time"
)

const drainWorkers = 8

type (
	TimingWheel struct {
		interval      time.Duration
		ticker        timex.Ticker
		slots         []*list.List
		timers        *SafeMap
		tickedPos     int
		numSlots      int
		execute       Execute
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

func newTimingWheelWithClock(interval time.Duration, numSlots int, execute Execute, ticker timex.Ticker) (*TimingWheel, error) {
	w := &TimingWheel{
		interval:      interval,
		ticker:        ticker,
		slots:         make([]*list.List, numSlots),
		timers:        NewSafeMap(),
		tickedPos:     numSlots - 1, // 位于上一次虚拟circle中
		numSlots:      numSlots,
		execute:       execute,
		setChannel:    make(chan timingEntry),
		moveChannel:   make(chan baseEntry),
		removeChannel: make(chan interface{}),
		drainChannel:  make(chan func(key, value interface{})),
		stopChannel:   make(chan lang.PlaceholderType),
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
	fmt.Println("初始化时间轮插槽")

	for i := 0; i < w.numSlots; i++ {
		w.slots[i] = list.New()
	}
}

func (w *TimingWheel) run() {
	fmt.Println("开始运行时间轮")

	for {
		select {
		case <-w.ticker.Chan():
			w.onTick()
		case task := <-w.setChannel:
			w.setTask(&task)
		case key := <-w.removeChannel:
			w.removeTask(key)
		case fn := <-w.drainChannel:
			w.drainAll(fn)
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
	val, ok := w.timers.Get(task.key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	if task.delay < w.interval {
		threading.RunSafe(func() {
			w.execute(timer.item.key, timer.item.value)
		})
		return
	}

	pos, circle := w.getPositionAndCircle(task.delay)
	if pos >= timer.pos {
		timer.item.circle = circle
		timer.item.diff = pos - timer.pos
	} else if circle > 0 {
		circle--
		timer.item.circle = circle
		timer.item.diff = w.numSlots + pos - timer.pos
	} else {
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
