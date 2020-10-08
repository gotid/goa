package cache

import "time"

type (
	Cache interface {
		Del(keys ...string) error
		Get(key string, dest interface{}) error
		Set(key string, value interface{}) error
		SetEx(key string, value interface{}, expiration time.Duration) error
		Take(dest interface{}, key string, query func(v interface{}) error) error
		TakeEx(dest interface{}, key string, query func(v interface{}) error) error
	}

	cacheCluster struct {
	}
)

func (cc cacheCluster) Del(keys ...string) error {
	panic("implement me")
}

func (cc cacheCluster) Get(key string, dest interface{}) error {
	panic("implement me")
}

func (cc cacheCluster) Set(key string, value interface{}) error {
	panic("implement me")
}

func (cc cacheCluster) SetEx(key string, value interface{}, expiration time.Duration) error {
	panic("implement me")
}

func (cc cacheCluster) Take(dest interface{}, key string, query func(v interface{}) error) error {
	panic("implement me")
}

func (cc cacheCluster) TakeEx(dest interface{}, key string, query func(v interface{}) error) error {
	panic("implement me")
}

func NewCache() Cache {
	return cacheCluster{}
}
