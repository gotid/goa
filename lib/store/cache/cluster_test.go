package cache

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"goa/lib/errorx"
	"goa/lib/hash"
	"goa/lib/store/redis"
	"goa/lib/syncx"
	"math"
	"strconv"
	"testing"
	"time"
)

type mockedNode struct {
	vals        map[string][]byte
	errNotFound error
}

func (n *mockedNode) Del(keys ...string) error {
	var es errorx.Errors
	for _, key := range keys {
		if _, ok := n.vals[key]; !ok {
			es.Add(n.errNotFound)
		} else {
			delete(n.vals, key)
		}
	}
	return es.Error()
}

func (n *mockedNode) Get(key string, dest interface{}) error {
	if bytes, ok := n.vals[key]; ok {
		return json.Unmarshal(bytes, dest)
	} else {
		return n.errNotFound
	}
}

func (n *mockedNode) Set(key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	n.vals[key] = data
	return nil
}

func (n *mockedNode) SetEx(key string, val interface{}, expires time.Duration) error {
	return n.Set(key, val)
}

func (n *mockedNode) Take(dest interface{}, key string, queryFn func(newVal interface{}) error) error {
	if _, ok := n.vals[key]; ok {
		return n.Get(key, dest)
	}

	if err := queryFn(dest); err != nil {
		return err
	}

	return n.Set(key, dest)
}

func (n *mockedNode) TakeEx(dest interface{}, key string, queryFn func(newVal interface{}, expires time.Duration) error) error {
	return n.Take(dest, key, func(newVal interface{}) error {
		return queryFn(newVal, 0)
	})
}

func TestCluster_SetDel(t *testing.T) {
	const total = 1000
	//r1 := miniredis.NewMiniRedis()
	//assert.Nil(t, r1.Start())
	//fmt.Println(r1.Addr())
	//defer r1.Close()
	//
	//r2 := miniredis.NewMiniRedis()
	//assert.Nil(t, r2.Start())
	//fmt.Println(r2.Addr())
	//defer r1.Close()

	confs := ClusterConf{
		{
			Conf: redis.Conf{
				Host: "127.0.0.1:6379",
				Mode: redis.StandaloneMode,
			},
			Weight: 100,
		},
		{
			Conf: redis.Conf{
				Host: "192.168.0.166:6800",
				Mode: redis.StandaloneMode,
			},
			Weight: 100,
		},
	}

	c := NewCacheCluster(confs, syncx.NewSharedCalls(), NewCacheStat("mock"), errPlaceholder)

	// 写
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			assert.Nil(t, c.Set(fmt.Sprintf("key/%d", i), i))
		} else {
			assert.Nil(t, c.SetEx(fmt.Sprintf("key/%d", i), i, 0))
		}
	}

	// 读
	for i := 0; i < total; i++ {
		var v int
		assert.Nil(t, c.Get(fmt.Sprintf("key/%d", i), &v))
		assert.Equal(t, i, v)
	}

	// 删
	for i := 0; i < total; i++ {
		assert.Nil(t, c.Del(fmt.Sprintf("key/%d", i)))
	}

	// 再读
	for i := 0; i < total; i++ {
		var v int
		assert.Equal(t, errPlaceholder, c.Get(fmt.Sprintf("key/%d", i), &v))
		assert.Equal(t, 0, v)
	}
}

func TestCluster_Balance(t *testing.T) {
	const (
		numNodes = 100
		total    = 10000
	)
	dispatcher := hash.NewConsistentHash()
	maps := make([]map[string][]byte, numNodes)

	// 初始值
	for i := 0; i < numNodes; i++ {
		maps[i] = map[string][]byte{
			strconv.Itoa(i): []byte(strconv.Itoa(i)),
		}
	}

	// 初始带权重的节点
	for i := 0; i < numNodes; i++ {
		dispatcher.AddWithWeight(&mockedNode{
			vals:        maps[i],
			errNotFound: errPlaceholder,
		}, 100)
	}

	// 初始化缓存集群
	c := cluster{
		dispatcher:  dispatcher,
		errNotFound: errPlaceholder,
	}

	// 写测试
	for i := 0; i < total; i++ {
		assert.Nil(t, c.Set(strconv.Itoa(i), i))
	}

	// 熵测试
	counts := make(map[int]int)
	for i, m := range maps {
		counts[i] = len(m)
	}
	entropy := calcEntropy(counts, total)
	assert.True(t, len(counts) > 1)
	assert.True(t, entropy > .95, fmt.Sprintf("熵应大于 0.95，但得到 %.2f", entropy))

	for i := 0; i < total; i++ {
		var v int
		assert.Nil(t, c.Get(strconv.Itoa(i), &v))
		assert.Equal(t, i, v)
	}

	for i := 0; i < total/10; i++ {
		assert.Nil(t, c.Del(strconv.Itoa(i*10), strconv.Itoa(i*10+1), strconv.Itoa(i*10+2)))
		assert.Nil(t, c.Del(strconv.Itoa(i*10+9)))
	}

	var count int
	for i := 0; i < total/10; i++ {
		var val int
		if i%2 == 0 {
			assert.Nil(t, c.Take(&val, strconv.Itoa(i*10), func(v interface{}) error {
				*v.(*int) = i
				count++
				return nil
			}))
		} else {
			assert.Nil(t, c.TakeEx(&val, strconv.Itoa(i*10), func(v interface{}, expire time.Duration) error {
				*v.(*int) = i
				count++
				return nil
			}))
		}
		assert.Equal(t, i, val)
	}
	assert.Equal(t, total/10, count)
}

func calcEntropy(m map[int]int, total int) float64 {
	var entropy float64

	for _, v := range m {
		proba := float64(v) / float64(total)
		entropy -= proba * math.Log2(proba)
	}

	return entropy / math.Log2(float64(len(m)))
}
