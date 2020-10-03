package lib

import (
	"goa/lib/errorx"
	"goa/lib/syncx"
	"io"
	"sync"
)

// ResourceManager 资源管理器提供可复用的资源
type ResourceManager struct {
	resources   map[string]io.Closer
	cachedCalls syncx.CachedCalls
	lock        sync.RWMutex
}

// NewResourceManager 返回可复用资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:   make(map[string]io.Closer),
		cachedCalls: syncx.NewCachedCalls(),
	}
}

func (m *ResourceManager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var errs errorx.Errors
	for _, res := range m.resources {
		if err := res.Close(); err != nil {
			errs.Append(err)
		}
	}
	return errs.Error()
}

func (m *ResourceManager) Get(key string, getFn func() (io.Closer, error)) (io.Closer, error) {
	result, _, err := m.cachedCalls.Get(key, func() (interface{}, error) {
		m.lock.Lock()
		res, ok := m.resources[key]
		m.lock.Unlock()
		if ok {
			return res, nil
		}

		res, err := getFn()
		if err != nil {
			return nil, err
		}

		m.lock.Lock()
		m.resources[key] = res
		m.lock.Unlock()
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(io.Closer), nil
}
