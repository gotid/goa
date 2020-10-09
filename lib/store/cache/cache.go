package cache

import (
	"fmt"
	"goa/lib/errorx"
	"goa/lib/hash"
	"goa/lib/logx"
	"goa/lib/syncx"
	"time"
)

type (
	Cache interface {
		Del(keys ...string) error
		Get(key string, dest interface{}) error
		Set(key string, value interface{}) error
		SetEx(key string, value interface{}, expires time.Duration) error
		Take(dest interface{}, key string, queryFn func(newVal interface{}) error) error
		TakeEx(dest interface{}, key string, queryFn func(newVal interface{}, expires time.Duration) error) error
	}

	cacheCluster struct {
		dispatcher  *hash.ConsistentHash
		errNotFound error
	}
)

func NewCache(confs ClusterConf, barrier syncx.SharedCalls, stat *Stat, errNotFound error, opts ...Option) Cache {
	if len(confs) == 0 || TotalWeights(confs) <= 0 {
		logx.Fatal("未配置缓存节点")
	}

	if len(confs) == 1 {
		return NewCacheNode(confs[0].NewRedis(), barrier, stat, errNotFound, opts...)
	}

	// 添加一批 redis 缓存节点
	dispatcher := hash.NewConsistentHash()
	for _, conf := range confs {
		node := NewCacheNode(conf.NewRedis(), barrier, stat, errNotFound, opts...)
		dispatcher.AddWithWeight(node, conf.Weight)
	}

	return cacheCluster{
		dispatcher:  dispatcher,
		errNotFound: errNotFound,
	}
}

func (cc cacheCluster) Del(keys ...string) error {
	switch len(keys) {
	case 0:
		return nil
	case 1:
		key := keys[0]
		c, ok := cc.dispatcher.Get(key)
		if !ok {
			return cc.errNotFound
		}
		return c.(Cache).Del(key)
	default:
		var es errorx.Errors
		nodes := make(map[interface{}][]string)
		for _, key := range keys {
			c, ok := cc.dispatcher.Get(key)
			if !ok {
				es.Add(fmt.Errorf("缓存 key %q 不存在", key))
				continue
			}

			nodes[c] = append(nodes[c], key)
		}
		for c, keys := range nodes {
			if err := c.(Cache).Del(keys...); err != nil {
				es.Add(err)
			}
		}

		return es.Error()
	}
}

func (cc cacheCluster) Get(key string, dest interface{}) error {
	c, ok := cc.dispatcher.Get(key)
	if !ok {
		return cc.errNotFound
	}

	return c.(Cache).Get(key, dest)
}

func (cc cacheCluster) Set(key string, value interface{}) error {
	c, ok := cc.dispatcher.Get(key)
	if !ok {
		return cc.errNotFound
	}

	return c.(Cache).Set(key, value)
}

func (cc cacheCluster) SetEx(key string, value interface{}, expires time.Duration) error {
	c, ok := cc.dispatcher.Get(key)
	if !ok {
		return cc.errNotFound
	}

	return c.(Cache).SetEx(key, value, expires)
}

func (cc cacheCluster) Take(dest interface{}, key string, queryFn func(v interface{}) error) error {
	c, ok := cc.dispatcher.Get(key)
	if !ok {
		return cc.errNotFound
	}

	return c.(Cache).Take(dest, key, queryFn)
}

func (cc cacheCluster) TakeEx(dest interface{}, key string, queryFn func(newVal interface{}, expires time.Duration) error) error {
	c, ok := cc.dispatcher.Get(key)
	if !ok {
		return cc.errNotFound
	}

	return c.(Cache).TakeEx(dest, key, queryFn)
}
