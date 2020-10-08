package redis

import (
	"fmt"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

func TestRedis_Exists(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := NewRedis(client.Addr, "").Exists("a")
		assert.NotNil(t, err)

		ok, err := client.Exists("a")
		assert.Nil(t, err)
		assert.False(t, ok)
		assert.Nil(t, client.Set("a", "b"))
		ok, err = client.Exists("a")
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Eval(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := NewRedis(client.Addr, "").Exists("a")
		assert.NotNil(t, err)

		_, err = client.Eval(`redis.call("EXISTS", KEYS[1])`, []string{"notexists"})
		assert.Equal(t, redis.Nil, err)

		err = client.Set("key1", "value1")
		assert.Nil(t, err)

		_, err = client.Eval(`redis.call("EXISTS", KEYS[1])`, []string{"key1"})
		assert.Equal(t, redis.Nil, err)

		val, err := client.Eval(`return redis.call("EXISTS", KEYS[1])`, []string{"key1"})
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
	})
}

func TestRedis_HGetAll(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))

		_, err := NewRedis(client.Addr, "").HGetAll("a")
		assert.NotNil(t, err)

		result, err := client.HGetAll("a")
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}, result)
	})
}

func TestRedis_HVals(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))

		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_HSetNX(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))

		ok, err := client.HSetNX("a", "bb", "ccc")
		assert.Nil(t, err)
		assert.False(t, ok)

		ok, err = client.HSetNX("a", "cc", "ccc")
		assert.Nil(t, err)
		assert.True(t, ok)

		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb", "ccc"}, vals)
	})
}

func TestRedis_HDelHLen(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))

		size, err := client.HLen("a")
		assert.Nil(t, err)
		assert.Equal(t, 2, size)

		ok, err := client.HDel("a", "aa")
		assert.Nil(t, err)
		assert.True(t, ok)

		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"bbb"}, vals)
	})
}

func TestRedis_HIncrBy(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		result, err := client.HIncrBy("key", "field", 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, result)

		result, err = client.HIncrBy("key", "field", 3)
		assert.Nil(t, err)
		assert.Equal(t, 5, result)
	})
}

func TestRedis_HKeys(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))

		keys, err := client.HKeys("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aa", "bb"}, keys)
	})
}

func TestRedis_HMGet(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))

		values, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "bbb"}, values)

		values, err = client.HMGet("a", "aa", "不存在的field", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "", "bbb"}, values)
	})
}

func TestRedis_HMSet(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HMSet("a", map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}))

		values, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb"}, values)
	})
}

func TestRedis_Incr(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		val, err := client.Incr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)

		val, err = client.Incr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
	})
}

func TestRedis_IncrBy(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := NewRedis(client.Addr, "").IncrBy("a", 2)
		assert.NotNil(t, err)

		val, err := client.IncrBy("a", 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)

		val, err = client.IncrBy("a", 3)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
	})
}

func TestRedis_Keys(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "value1")
		assert.Nil(t, err)
		err = client.Set("key2", "value2")
		assert.Nil(t, err)
		_, err = NewRedis(client.Addr, "").Keys("*")
		assert.NotNil(t, err)
		keys, err := client.Keys("*")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})
}

func TestRedis_HyperLogLog(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()

		r := NewRedis(client.Addr, "")
		_, err := r.PFAdd("key1", "a", "b")
		assert.NotNil(t, err)

		_, err = r.PFCount("*")
		assert.NotNil(t, err)

		err = r.PFMerge("*")
		assert.NotNil(t, err)
	})
}

func TestRedis_List(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		length, err := client.LPush("key", "value1", "value2")
		assert.Nil(t, err)
		assert.Equal(t, 2, length)

		length, err = client.RPush("key", "value3", "value4")
		assert.Nil(t, err)
		assert.Equal(t, 4, length)

		length, err = client.LLen("key")
		assert.Nil(t, err)
		assert.Equal(t, 4, length)

		values, err := client.LRange("key", 0, -1)
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"value2", "value1", "value3", "value4"}, values)

		val, err := client.LPop("key")
		assert.Nil(t, err)
		assert.Equal(t, "value2", val)
	})
}

func TestRedis_MGet(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.Set("key1", "value1"))
		assert.Nil(t, client.Set("key2", "value2"))

		values, err := client.MGet("key1", "key2")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value1", "value2"}, values)
	})
}

func TestRedis_SetBit(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.SetBit("key", 1, 1))
	})
}

func TestRedis_GetBit(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.SetBit("key", 2, 1))

		bit, err := client.GetBit("key", 2)
		assert.Nil(t, err)
		assert.Equal(t, 1, bit)
	})
}

func TestRedis_Persist(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		ok, err := client.Persist("key")
		assert.Nil(t, err)
		assert.False(t, ok)

		assert.Nil(t, client.Set("key", "value"))
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.False(t, ok)

		assert.Nil(t, client.Expire("key", 5))
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.True(t, ok)

		assert.Nil(t, client.ExpireAt("key", time.Now().Unix()+5))
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Ping(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		ok := client.Ping()
		assert.True(t, ok)
	})
}

func TestRedis_Scan(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.Set("key1", "value1"))
		assert.Nil(t, client.Set("key2", "value2"))
		assert.Nil(t, client.Set("a1", "value2"))
		assert.Nil(t, client.Set("a2", "value2"))

		keys, _, err := client.Scan(0, "*1", 100)
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"key1", "a1"}, keys)
	})
}

func TestRedis_Set(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		var list []string
		for i := 0; i < 1500; i++ {
			list = append(list, fmt.Sprintf("value_%d", i))
		}
		length, err := client.SAdd("set", list)
		assert.Nil(t, err)
		assert.Equal(t, 1500, length)

		var cursor uint64 = 0
		var total = 0
		for {
			keys, nextCur, err := client.SScan("set", cursor, "", 100)
			assert.Nil(t, err)
			total += len(keys)
			if nextCur == 0 {
				break
			}
			cursor = nextCur
		}
		assert.Equal(t, 1500, total)

		card, err := client.SCard("set")
		assert.Nil(t, err)
		assert.Equal(t, int64(1500), card)

		ok, err := client.SIsMember("set", "value_2")
		assert.Nil(t, err)
		assert.True(t, ok)

		length, err = client.SRem("set", "value_2", "value_3")
		assert.Nil(t, err)
		assert.Equal(t, 2, length)

		members, err := client.SRandMemberN("set", 3)
		assert.Nil(t, err)
		fmt.Println(members)

		length, err = client.Del("set")
		assert.Nil(t, err)
		assert.Equal(t, 1, length)
	})
}

func runOnRedis(t *testing.T, fn func(client *Redis)) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	defer func() {
		client, err := standalonClientManager.Get(s.Addr(), func() (io.Closer, error) {
			//return nil, errors.New("可能已经存在")
			return nil, nil
		})
		if err != nil {
			t.Error(err)
		}

		if client != nil {
			client.Close()
		}
	}()

	fn(NewRedis(s.Addr(), StandaloneMode))
}
