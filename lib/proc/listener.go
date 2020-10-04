package proc

import "sync"

type ListenerManager struct {
	lock      sync.Mutex
	wg        sync.WaitGroup
	listeners []func()
}

func (m *ListenerManager) add(fn func()) (waitForCalled func()) {
	m.wg.Add(1)

	m.lock.Lock()
	m.listeners = append(m.listeners, func() {
		defer m.wg.Done()
		fn()
	})
	m.lock.Unlock()

	return func() {
		m.wg.Wait()
	}
}

func (m *ListenerManager) notify() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, listener := range m.listeners {
		listener()
	}
}
