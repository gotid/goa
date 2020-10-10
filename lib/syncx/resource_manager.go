package syncx

import (
	"github.com/z-sdk/goa/lib/errorx"
	"io"
	"sync"
)

// ResourceManager 资源管理器提供可复用的资源
type ResourceManager struct {
	resources   map[string]io.Closer
	cachedCalls SharedCalls
	lock        sync.RWMutex
}

// NewResourceManager 返回可复用资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:   make(map[string]io.Closer),
		cachedCalls: NewSharedCalls(),
	}
}

// Close 关闭该资源管理器。
func (m *ResourceManager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var errs errorx.Errors
	for _, res := range m.resources {
		if err := res.Close(); err != nil {
			errs.Add(err)
		}
	}
	return errs.Error()
}

// Get 获取指定 key 的缓存资源，如缓存不存在，则调用 getFn 回源获取并加入缓存资源映射。
func (m *ResourceManager) Get(key string, getFn func() (io.Closer, error)) (io.Closer, error) {
	result, _, err := m.cachedCalls.Do(key, func() (interface{}, error) {
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
