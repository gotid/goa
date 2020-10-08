package redis

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"goa/lib/mapping"
	"strconv"
	"time"
)

const (
	blockingTimeout = 5 * time.Second
)

var ErrNilConn = errors.New("redis 连接不可为空")

// BLPop 阻塞式列表弹出操作
func (r *Redis) BLPop(conn Client, key string) (string, error) {
	if conn == nil {
		return "", ErrNilConn
	}

	result, err := conn.BLPop(blockingTimeout, key).Result()
	if err != nil {
		return "", err
	}

	if len(result) < 2 {
		return "", fmt.Errorf("键上无值：%s", key)
	} else {
		return result[1], nil
	}
}

// Del 时间复杂度 O(N)，当删除的key是字符串意外的复杂类型如List、Set、Hash等则为 O(1)
func (r *Redis) Del(keys ...string) (result int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.Del(keys...).Result()
		if err != nil {
			return err
		}
		result = int(v)
		return nil
	}, acceptable)
	return
}

// Eval 求解 Lua 脚本
func (r *Redis) Eval(script string, keys []string, args ...interface{}) (result interface{}, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		result, err = client.Eval(script, keys, args...).Result()
		return err
	}, acceptable)

	return
}

// Exists 判断 key 是否存在
func (r *Redis) Exists(key string) (result bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.Exists(key).Result()
		if err != nil {
			return err
		}
		result = v == 1 // 1 存在，0 不存在
		return nil
	}, acceptable)

	return
}

// Expire 设置 key 的过期秒数
func (r *Redis) Expire(key string, seconds int) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		return client.Expire(key, time.Duration(seconds)*time.Second).Err()
	}, acceptable)
}

// ExpireAt 设置 key 的过期时间（单位：秒）
func (r *Redis) ExpireAt(key string, expireTime int64) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		return client.ExpireAt(key, time.Unix(expireTime, 0)).Err()
	}, acceptable)
}

// Get 获取 key 对应的字符串值
func (r *Redis) Get(key string) (result string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		result, err = client.Get(key).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		return nil
	}, acceptable)

	return
}

// GetBit 返回 key 对应字符串在 offset 处的 bit 值。不存在则返回 0。
func (r *Redis) GetBit(key string, offset int64) (result int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.GetBit(key, offset).Result()
		if err != nil {
			return err
		}
		result = int(v)
		return nil
	}, acceptable)

	return
}

// HDel 从 key 指定的哈希集合中删除指定 field。成功返回1，失败返回0。
func (r *Redis) HDel(key, field string) (result bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.HDel(key, field).Result()
		if err != nil {
			return err
		}
		result = v == 1
		return nil
	}, acceptable)

	return
}

// HExists 判断指定 key 的哈希中 field 是否存在，存在返1，否则返0
func (r *Redis) HExists(key, field string) (result bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}
		result, err = client.HExists(key, field).Result()
		return err
	}, acceptable)

	return
}

// HExists 从指定 key 的哈希中指定 field 的字符串值。
func (r *Redis) HGet(key, field string) (result string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}
		result, err = client.HGet(key, field).Result()
		return err
	}, acceptable)

	return
}

// HGetAll 获取指定 key 的哈希中所有键值对
func (r *Redis) HGetAll(key string) (result map[string]string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		result, err = client.HGetAll(key).Result()
		return err
	}, acceptable)

	return
}

// HIncrBy 增加 key 指定的哈希集中指定字段的数值。
func (r *Redis) HIncrBy(key, field string, increment int) (result int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.HIncrBy(key, field, int64(increment)).Result()
		if err != nil {
			return err
		}
		result = int(v)
		return nil
	}, acceptable)

	return
}

// HKeys 返回 key 指定的哈希集中所有字段的名字。
func (r *Redis) HKeys(key string) (fields []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		fields, err = client.HKeys(key).Result()
		return err
	}, acceptable)

	return
}

// HLen 返回 key 指定的哈希的字段数量。
func (r *Redis) Hlen(key string) (size int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.HLen(key).Result()
		if err != nil {
			return err
		}
		size = int(v)
		return nil
	}, acceptable)

	return
}

// HMGet 返回 key 指定的哈希集中指定字段的值。
func (r *Redis) HMGet(key string, fields ...string) (values []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.HMGet(key, fields...).Result()
		if err != nil {
			return err
		}
		values = toStrings(v)
		return nil
	}, acceptable)

	return
}

// HSet 设置 key 指定的哈希集中指定字段的值。
func (r *Redis) HSet(key, field, value string) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		return client.HSet(key, field, value).Err()
	}, acceptable)
}

// HSetNX 只在 key 指定的哈希集中不存在指定的字段时，设置字段的值。
func (r *Redis) HSetNX(key, field, value string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.HSetNX(key, field, value).Result()
		return err
	}, acceptable)

	return
}

// HMSet 设置 key 指定的哈希集中指定字段的值。
// 该命令将重写所有在哈希集中存在的字段。
// 如果 key 指定的哈希集不存在，会创建一个新的哈希集并与 key 关联。
func (r *Redis) HMSet(key string, fieldsAndValues map[string]string) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		vals := make(map[string]interface{}, len(fieldsAndValues))
		for k, v := range fieldsAndValues {
			vals[k] = v
		}

		return client.HMSet(key, vals).Err()
	}, acceptable)
}

// HVals 返回指定 key 哈希中的值列表
func (r *Redis) HVals(key string) (values []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		values, err = client.HVals(key).Result()
		return err
	}, acceptable)

	return
}

// Incr 增加 key 对应的数值。
func (r *Redis) Incr(key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.Incr(key).Result()
		return err
	}, acceptable)

	return
}

func (r *Redis) IncrBy(key string, increment int64) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.IncrBy(key, increment).Result()
		return err
	}, acceptable)

	return
}

// Keys 查找所有符合给定模式 pattern（正则表达式）的 key 列表。
func (r *Redis) Keys(pattern string) (keys []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		keys, err = client.Keys(pattern).Result()
		return err
	}, acceptable)

	return
}

// LLen 返回指定 key 的列表长度。
func (r *Redis) LLen(key string) (length int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.LLen(key).Result()
		if err != nil {
			return err
		}
		length = int(v)
		return nil
	}, acceptable)

	return
}

// LPop 移除并返回 key 对应的 list 的第一个元素。
func (r *Redis) LPop(key string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.LPop(key).Result()
		return err
	}, acceptable)

	return
}

// LPush 将所有指定的值插入到 key 对应列表的头部。列表不存在会自动创建。
func (r *Redis) LPush(key string, values ...interface{}) (length int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.LPush(key, values...).Result()
		if err != nil {
			return err
		}
		length = int(v)
		return nil
	}, acceptable)

	return
}

// LRange 返回存储在 key 的列表里指定范围内的元素。
// 偏移量也可以是负数，表示偏移量是从list尾部开始计数。
func (r *Redis) LRange(key string, start int, stop int) (values []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		values, err = client.LRange(key, int64(start), int64(stop)).Result()
		return err
	}, acceptable)

	return
}

// LRem
//
// 从存于 key 的列表里移除前 count 次出现的值为 value 的元素。 这个 count 参数通过下面几种方式影响这个操作：
//
// count > 0: 从头往尾移除值为 value 的元素。
//
// count < 0: 从尾往头移除值为 value 的元素。
//
// count = 0: 移除所有值为 value 的元素。
//
// 比如， LREM list -2 “hello” 会从存于 list 的列表里移除最后两个出现的 “hello”。
func (r *Redis) LRem(key string, count int, value string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.LRem(key, int64(count), value).Result()
		if err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// MGet 返回所有指定的key的value。
//
// 对于每个不对应string或者不存在的key，都返回特殊值nil。
// 正因为此，这个操作从来不会失败。
func (r *Redis) MGet(keys ...string) (values []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.MGet(keys...).Result()
		if err != nil {
			return err
		} else {
			values = toStrings(v)
			return nil
		}
	}, acceptable)

	return
}

// Persist 移除给定key的生存时间
//
// 将这个 key 从『易失的』(带生存时间 key )转换成『持久的』(一个不带生存时间、永不过期的 key )。
func (r *Redis) Persist(key string) (result bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		result, err = client.Persist(key).Result()
		return err
	}, acceptable)

	return
}

// PFAdd 将除了第一个参数以外的参数存储到以第一个参数为变量名的HyperLogLog结构中.
func (r *Redis) PFAdd(key string, values ...interface{}) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.PFAdd(key, values...).Result(); err != nil {
			return err
		} else {
			val = v == 1
			return nil
		}
	}, acceptable)

	return
}

// PFCount 返回存储在HyperLogLog结构体的该变量的近似基数，如果该变量不存在,则返回0.
func (r *Redis) PFCount(key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.PFCount(key).Result()
		return err
	}, acceptable)

	return
}

// PFMerge 将多个 HyperLogLog 合并（merge）为一个 HyperLogLog
//
// 合并后的 HyperLogLog 的基数接近于所有输入 HyperLogLog 的可见集合（observed set）的并集.
func (r *Redis) PFMerge(dest string, keys ...string) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		_, err = client.PFMerge(dest, keys...).Result()
		return err
	}, acceptable)
}

// Ping 测试连接是否可用
func (r *Redis) Ping() (result bool) {
	_ = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.Ping().Result(); err != nil {
			result = false
			return nil
		} else {
			result = v == "PONG"
			return nil
		}
	}, acceptable)

	return
}

// Pipelined 运行指定的 fn 管道处理函数
func (r *Redis) Pipelined(fn func(Pipeliner) error) (err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		_, err = client.Pipelined(fn)
		return err

	}, acceptable)

	return
}

// RPush 从右侧向 key 对应列表中插入一组值
func (r *Redis) RPush(key string, values ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.RPush(key, values...).Result(); err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// SAdd 添加一个或多个指定的member元素到集合的 key 中。
func (r *Redis) SAdd(key string, values ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.SAdd(key, values...).Result(); err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// Scan 命令是一个基于游标的迭代器，用于迭代当前数据库中的key集合。
//
// cursor 迭代起始游标
//
// match 正则匹配模式
//
// count 为此次迭代期望返回条的数
func (r *Redis) Scan(cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		keys, cur, err = client.Scan(cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

// SetBit 设置或者清空key的value(字符串)在offset处的bit值。
func (r *Redis) SetBit(key string, offset int64, value int) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		_, err = client.SetBit(key, offset, value).Result()
		return err
	}, acceptable)
}

// SScan 命令是一个基于游标的迭代器，用于迭代当前 key 集合。
//
// cursor 迭代起始游标
//
// match 正则匹配模式
//
// count 为此次迭代期望返回条的数
func (r *Redis) SScan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		keys, cur, err = client.SScan(key, cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

// SCard 返回集合存储的key的基数 (集合元素的数量)。
func (r *Redis) SCard(key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SCard(key).Result()
		return err
	}, acceptable)

	return
}

// Set 将键key设定为指定的“字符串”值。
func (r *Redis) Set(key string, value string) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		return client.Set(key, value, 0).Err()
	}, acceptable)
}

// SetEx 设置key对应字符串value，并且设置key在给定的seconds时间之后超时过期。
//
// seconds 多少秒后过期，0 代表永不过期。
func (r *Redis) SetEx(key, value string, seconds int) error {
	return r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		return client.Set(key, value, time.Duration(seconds)*time.Second).Err()
	}, acceptable)
}

// SetNX 将 key 设置值为 value。
//
// 如果key不存在，这种情况下等同SET命令。
//
// 当key存在时，什么也不做。SETNX是”SET if Not eXists”的简写。
func (r *Redis) SetNX(key, value string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SetNX(key, value, 0).Result()
		return err
	}, acceptable)

	return
}

// SetNXEx 设置一个不存在的key值为value，并设置过期时间。如果key存在则什么也不做。
func (r *Redis) SetNXEx(key, value string, seconds int) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SetNX(key, value, time.Duration(seconds)*time.Second).Result()
		return err
	}, acceptable)

	return
}

// SIsMemeber 返回成员 member 是否是存储的集合 key 的成员.
func (r *Redis) SIsMember(key string, member interface{}) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}
		val, err = client.SIsMember(key, member).Result()
		return err
	}, acceptable)

	return
}

// SRem 在key集合中移除指定的元素.
func (r *Redis) SRem(key string, members ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.SRem(key, members...).Result(); err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// SMembers 返回 key 集合中的成员列表
func (r *Redis) SMembers(key string) (members []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		members, err = client.SMembers(key).Result()
		return err
	}, acceptable)

	return
}

// SPop 从存储在key的集合中【移除并返回】一个或多个【随机】元素。
func (r *Redis) SPop(key string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SPop(key).Result()
		return err
	}, acceptable)

	return
}

// SRandMemberN 它从一个集合中返回N个随机元素，但不删除元素。
func (r *Redis) SRandMemberN(key string, count int) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SRandMemberN(key, int64(count)).Result()
		return err
	}, acceptable)

	return
}

// SUnion 返回给定的多个集合的并集中的所有成员.
func (r *Redis) SUnion(keys ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SUnion(keys...).Result()
		return err
	}, acceptable)

	return
}

// SUnionStore 类似于SUNION命令,不同的是它并不返回结果集,而是将结果存储在destination集合中.
//
// 如果destination 已经存在,则将其覆盖.
func (r *Redis) SUnionStore(destination string, keys ...string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.SUnionStore(destination, keys...).Result(); err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// SDiff 返回一个集合与给定集合的差集的元素。
func (r *Redis) SDiff(keys ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		val, err = client.SDiff(keys...).Result()
		return err
	}, acceptable)

	return
}

// SDiffStore 类似于 SDIFF, 不同之处在于该命令不返回结果集，而是将结果存放在destination集合中.
func (r *Redis) SDiffStore(destination string, keys ...string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.SDiffStore(destination, keys...).Result(); err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// TTL 返回 key 剩余的过期时间。
func (r *Redis) TTL(key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if duration, err := client.TTL(key).Result(); err != nil {
			return err
		} else {
			val = int(duration / time.Second)
			return nil
		}
	}, acceptable)

	return
}

// ZAdd 将所有指定成员添加到键为key有序集合（sorted set）里面。
//
// key 有序集合的key
//
// store 成员排序分值
//
// member 成员名称
func (r *Redis) ZAdd(key string, score int64, member string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZAdd(key, redis.Z{
			Score:  float64(score),
			Member: member,
		}).Result(); err != nil {
			return err
		} else {
			val = v == 1
			return nil
		}
	}, acceptable)

	return
}

// ZAdds 添加一组成员到键为key有序集合（sorted set）里面。
func (r *Redis) ZAdds(key string, pairs ...Pair) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		var zs []redis.Z
		for _, pair := range pairs {
			z := redis.Z{Score: float64(pair.Score), Member: pair.Key}
			zs = append(zs, z)
		}

		if v, err := client.ZAdd(key, zs...).Result(); err != nil {
			return err
		} else {
			val = v
			return nil
		}
	}, acceptable)

	return
}

// ZCard 返回key的有序集元素个数。
func (r *Redis) ZCard(key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZCard(key).Result(); err != nil {
			return err
		} else {
			val = int(v)
			return nil
		}
	}, acceptable)

	return
}

// ZCount 返回有序集key中，score值在min和max之间(默认包括score值等于min或max)的成员个数。
func (r *Redis) ZCount(key string, min, max int64) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZCount(key,
			strconv.FormatInt(min, 10),
			strconv.FormatInt(max, 10)).Result()
		if err != nil {
			return err
		}
		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZIncrBy 为有序集key的成员member的score值加上增量increment。
//
// 如果key中不存在member，就在key中添加一个member，score是increment（就好像它之前的score是0.0）。
//
// 如果key不存在，就创建一个只含有指定member成员的有序集合。
func (r *Redis) ZIncrBy(key string, increment int64, member string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZIncrBy(key, float64(increment), member).Result(); err != nil {
			return err
		} else {
			val = int64(v)
			return nil
		}
	}, acceptable)

	return
}

// ZScore 返回有序集key中，成员member的score值。
func (r *Redis) ZScore(key string, member string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZScore(key, member).Result(); err != nil {
			return err
		} else {
			val = int64(v)
			return nil
		}
	}, acceptable)

	return
}

// ZRank 返回有序集key中成员member的排名。

// 其中有序集成员按score值递增(从小到大)顺序排列。

// 排名以0为底，也就是说，score值最小的成员排名为0。
// 倒序排名，使用 ZREVRANK
func (r *Redis) ZRank(key, member string) (rank int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		rank, err = client.ZRank(key, member).Result()
		return err
	}, acceptable)

	return
}

// ZRem 从 key 有序集中删除指定的 members 成员，返回删除个数。
func (r *Redis) ZRem(key string, members ...interface{}) (num int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZRem(key, members...).Result(); err != nil {
			return err
		} else {
			num = int(v)
			return nil
		}
	}, acceptable)

	return
}

// ZRemRangeByScore 移除有序集key中，所有score值介于min和max之间(包括等于min或max)的成员。
//
// 自版本2.1.6开始，score值等于min或max的成员也可以不包括在内，详见 -inf +inf 和 (
// http://www.redis.cn/commands/zrangebyscore.html
func (r *Redis) ZRemRangeByScore(key string, min, max int64) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZRemRangeByScore(key,
			strconv.FormatInt(min, 10),
			strconv.FormatInt(max, 10)).Result()
		if err != nil {
			return err
		}
		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZRemRangeByRank 移除有序集key中，指定排名(rank)区间内的所有成员。
//
// 下标参数start和stop都以0为底，0处是分数最小的那个元素。
//
// 这些索引也可是负数，表示位移从最高分处开始数。例如，-1是分数最高的元素，-2是分数第二高的，依次类推。
func (r *Redis) ZRemRangeByRank(key string, start, stop int64) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZRemRangeByRank(key, start, stop).Result()
		if err != nil {
			return err
		}
		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZRange 返回存储在有序集合key中的指定范围的元素。
//
// 返回的元素可以认为是按得分从最低到最高排列。 如果得分相同，将按字典排序。
//
// 当你需要元素从最高分到最低分排列时，请参阅 ZREVRANGE
func (r *Redis) ZRange(key string, start, stop int64) (result []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		result, err = client.ZRange(key, start, stop).Result()
		return err
	}, acceptable)

	return
}

// ZRangeWithScores 返回key的有序集合中的分数在min和max之间的所有元素（包括分数等于max或者min的元素）。
// 元素被认为是从低分到高分排序的。
func (r *Redis) ZRangeWithScores(key string, start, stop int64) (pairs []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZRangeWithScores(key, start, stop).Result(); err != nil {
			return err
		} else {
			pairs = toPairs(v)
			return nil
		}
	}, acceptable)

	return
}

// ZRevRangeWithScores 返回有序集合中指定分数区间内的成员，分数由高到低排序。
func (r *Redis) ZRevRangeWithScores(key string, start, stop int64) (pairs []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		if v, err := client.ZRevRangeWithScores(key, start, stop).Result(); err != nil {
			return err
		} else {
			pairs = toPairs(v)
			return nil
		}
	}, acceptable)

	return
}

// ZRangeByScoreWithScore 按得分升序取key有序集成员
//
// - 返回数据带得分
func (r *Redis) ZRangeByScoreWithScores(key string, start, stop int64) (pairs []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZRangeByScoreWithScores(key,
			redis.ZRangeBy{
				Min: strconv.FormatInt(start, 10),
				Max: strconv.FormatInt(stop, 10),
			}).Result()
		if err != nil {
			return err
		}
		pairs = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRangeByScoreWithScoresAndLimit 按得分升序取key有序集成员
//
// - 返回数据带得分
//
// - 支持分页
func (r *Redis) ZRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	pairs []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		if size <= 0 {
			return nil
		}

		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZRangeByScoreWithScores(key, redis.ZRangeBy{
			Min:    strconv.FormatInt(start, 10),
			Max:    strconv.FormatInt(stop, 10),
			Offset: int64(page * size),
			Count:  int64(size),
		}).Result()
		if err != nil {
			return err
		} else {
			pairs = toPairs(v)
			return nil
		}
	}, acceptable)

	return
}

// ZRevRange 返回有序集key中，指定区间内的成员。
//
// - 返回数据带得分
func (r *Redis) ZRevRange(key string, start, stop int64) (result []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		result, err = client.ZRevRange(key, start, stop).Result()
		return err
	}, acceptable)

	return
}

// ZRevRangeByScoreWithScores 返回有序集合中指定分数区间内的成员，分数由高到低排序。
func (r *Redis) ZRevRangeByScoreWithScores(key string, start, stop int64) (pairs []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZRevRangeByScoreWithScores(key, redis.ZRangeBy{
			Min: strconv.FormatInt(start, 10),
			Max: strconv.FormatInt(stop, 10),
		}).Result()

		if err != nil {
			return err
		}
		pairs = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRevRangeByScoreWithScoresAndLimit 按得分降序取key有序集成员
//
// - 返回数据带得分
//
// - 支持分页
func (r *Redis) ZRevRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	pairs []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		if size <= 0 {
			return nil
		}

		client, err := getClient(r)
		if err != nil {
			return err
		}

		v, err := client.ZRevRangeByScoreWithScores(key, redis.ZRangeBy{
			Min:    strconv.FormatInt(start, 10),
			Max:    strconv.FormatInt(stop, 10),
			Offset: int64(page * size),
			Count:  int64(size),
		}).Result()
		if err != nil {
			return err
		} else {
			pairs = toPairs(v)
			return nil
		}
	}, acceptable)

	return
}

// scriptLoad 加载 Lua 脚本
func (r *Redis) scriptLoad(script string) (result string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		client, err := getClient(r)
		if err != nil {
			return err
		}
		result, err = client.ScriptLoad(script).Result()
		return err
	}, acceptable)
	return
}

// 断路器判断错误是否可接受，进而决定accepts是否+1
func acceptable(err error) bool {
	return err == nil || err == redis.Nil
}

func toPairs(vals []redis.Z) []Pair {
	pairs := make([]Pair, len(vals))
	for i, val := range vals {
		switch member := val.Member.(type) {
		case string:
			pairs[i] = Pair{
				Key:   member,
				Score: int64(val.Score),
			}
		default:
			pairs[i] = Pair{
				Key:   mapping.Repr(val.Member),
				Score: int64(val.Score),
			}
		}
	}
	return pairs
}

func toStrings(vals []interface{}) []string {
	ret := make([]string, len(vals))
	for i, val := range vals {
		if val == nil {
			ret[i] = ""
		} else {
			switch val := val.(type) {
			case string:
				ret[i] = val
			default:
				ret[i] = mapping.Repr(val)
			}
		}
	}
	return ret
}
