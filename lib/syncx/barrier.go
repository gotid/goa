package syncx

import "sync"

// Barrier 屏障器：对操作加锁
type Barrier struct {
	lock sync.Mutex
}

// Guard 加锁保护
func (b *Barrier) Guard(fn func()) {
	b.lock.Lock()
	defer b.lock.Unlock()
	fn()
}
