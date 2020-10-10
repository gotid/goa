package cache

import (
	"fmt"
	"github.com/z-sdk/goa/lib/errorx"
	"github.com/z-sdk/goa/lib/hash"
	"github.com/z-sdk/goa/lib/logx"
	"github.com/z-sdk/goa/lib/syncx"
	"time"
)

type (
	Cache interface {
		Del(keys ...string) error
		Get(key string, dest interface{}) error
		Set(key string, val interface{}) error
		SetEx(key string, val interface{}, expires time.Duration) error
		Take(dest interface{}, key string, queryFn func(interface{}) error) error
		TakeEx(dest interface{}, key string, queryFn func(interface{}, time.Duration) error) error
	}

	cluster struct {
		dispatcher  *hash.ConsistentHash
		errNotFound error
	}
)

func NewCacheCluster(confs ClusterConf, barrier syncx.SharedCalls, stat *Stat, errNotFound error, opts ...Option) Cache {
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

	return cluster{
		dispatcher:  dispatcher,
		errNotFound: errNotFound,
	}
}

func (c cluster) Del(keys ...string) error {
	switch len(keys) {
	case 0:
		return nil
	case 1:
		key := keys[0]
		node, ok := c.dispatcher.Get(key)
		if !ok {
			return c.errNotFound
		}
		return node.(Cache).Del(key)
	default:
		var es errorx.Errors
		nodes := make(map[interface{}][]string)
		for _, key := range keys {
			node, ok := c.dispatcher.Get(key)
			if !ok {
				es.Add(fmt.Errorf("缓存 key %q 不存在", key))
				continue
			}

			nodes[node] = append(nodes[node], key)
		}
		for node, keys := range nodes {
			if err := node.(Cache).Del(keys...); err != nil {
				es.Add(err)
			}
		}

		return es.Error()
	}
}

func (c cluster) Get(key string, dest interface{}) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).Get(key, dest)
}

func (c cluster) Set(key string, value interface{}) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).Set(key, value)
}

func (c cluster) SetEx(key string, value interface{}, expires time.Duration) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).SetEx(key, value, expires)
}

func (c cluster) Take(dest interface{}, key string, queryFn func(v interface{}) error) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).Take(dest, key, queryFn)
}

func (c cluster) TakeEx(dest interface{}, key string, queryFn func(newVal interface{}, expires time.Duration) error) error {
	node, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return node.(Cache).TakeEx(dest, key, queryFn)
}
