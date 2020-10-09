package collection

import "sync"

const (
	maxDeleted    = 10000
	copyThreshold = 1000
)

type SafeMap struct {
	lock       sync.RWMutex
	deletedOld int
	deletedNew int
	dirtyOld   map[interface{}]interface{}
	dirtyNew   map[interface{}]interface{}
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		dirtyOld: make(map[interface{}]interface{}),
		dirtyNew: make(map[interface{}]interface{}),
	}
}

func (m *SafeMap) Del(key interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.dirtyOld[key]; ok {
		delete(m.dirtyOld, key)
		m.deletedOld++
	} else if _, ok := m.dirtyNew[key]; ok {
		delete(m.dirtyNew, key)
		m.deletedNew++
	}

	if m.deletedOld >= maxDeleted && len(m.dirtyOld) < copyThreshold {
		for k, v := range m.dirtyOld {
			m.dirtyNew[k] = v
		}
		m.dirtyOld = m.dirtyNew
		m.deletedOld = m.deletedNew
		m.dirtyNew = make(map[interface{}]interface{})
		m.deletedNew = 0
	}

	if m.deletedNew >= maxDeleted && len(m.dirtyNew) < copyThreshold {
		for k, v := range m.dirtyNew {
			m.dirtyOld[k] = v
		}
		m.dirtyNew = make(map[interface{}]interface{})
		m.deletedNew = 0
	}
}

func (m *SafeMap) Get(key interface{}) (interface{}, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if val, ok := m.dirtyOld[key]; ok {
		return val, true
	} else {
		val, ok := m.dirtyNew[key]
		return val, ok
	}
}

func (m *SafeMap) Set(key, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.deletedOld <= maxDeleted {
		if _, ok := m.dirtyNew[key]; ok {
			delete(m.dirtyNew, key)
			m.deletedNew++
		}
		m.dirtyOld[key] = value
	} else {
		if _, ok := m.dirtyOld[key]; ok {
			delete(m.dirtyOld, key)
			m.deletedOld++
		}
		m.dirtyNew[key] = value
	}
}

func (m *SafeMap) Size() int {
	m.lock.Lock()
	size := len(m.dirtyOld) + len(m.dirtyNew)
	m.lock.Unlock()
	return size
}
