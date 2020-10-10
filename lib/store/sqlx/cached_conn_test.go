package sqlx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"goa/lib/logx"
	"goa/lib/stat"
	"goa/lib/store/cache"
	"goa/lib/store/redis"
	"sync/atomic"
	"testing"
	"time"
)

var (
	cacheUserIdPrefix       = "cache#User#id#"
	cacheUserNicknamePrefix = "cache#User#nickname#"
)

type Profile struct {
	ID       int64  `db:"id"`
	Kind     int    `db:"kind"`
	Nickname string `db:"nickname"`
	TestId   int    `db:"test_id"`
}

func init() {
	logx.Disable()
	stat.SetReporter(nil)
}

func TestCachedConn_FindOne(t *testing.T) {
	resetStats()
	r := redis.NewRedis("192.168.0.166:6800", redis.StandaloneMode)
	conn := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_user?parseTime=true")
	c := NewCachedConn(conn, r, cache.WithExpires(time.Minute))

	userId := 128
	userIdKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, userId)
	var profile Profile
	var err error
	err = c.Query(&profile, userIdKey, func(conn Conn, dest interface{}) error {
		query := `select id, kind, nickname from nest_user.profile where id=? limit 1`
		return conn.Query(dest, query, userId)
	})
	fmt.Println(err)
	fmt.Println(profile)
	fmt.Println("Total", atomic.LoadUint64(&cacheStat.Total))
	fmt.Println("Hit", atomic.LoadUint64(&cacheStat.Hit))
	fmt.Println("Miss", atomic.LoadUint64(&cacheStat.Miss))
	fmt.Println("DbFails", atomic.LoadUint64(&cacheStat.DbFails))
}

func TestCachedConn_FindByIndex(t *testing.T) {
	resetStats()
	r := redis.NewRedis("192.168.0.166:6800", redis.StandaloneMode)
	conn := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_user?parseTime=true")
	//c := NewCachedConn(conn, r, cache.WithExpires(10*time.Second))
	c := NewCachedConn(conn, r) // 默认缓存7天

	var profile Profile
	nickname := "测试小号9"
	nicknameKey := fmt.Sprintf("%s%v", cacheUserNicknamePrefix, nickname)

	// 通过索引键直接获取缓存结果
	err := c.QueryIndex(&profile, nicknameKey, func(id interface{}) string {
		// 取主键缓存键
		return fmt.Sprintf("%s%v", cacheUserIdPrefix, id)
	}, func(conn Conn, dest interface{}) (interface{}, error) {
		fmt.Println("索引查询", nickname, dest)
		// 通过索引查行记录
		query := `select id, kind, nickname from nest_user.profile where nickname=?`
		if err := conn.Query(&profile, query, nickname); err != nil {
			return nil, err
		}
		return profile.ID, nil
	}, func(conn Conn, dest, id interface{}) error {
		fmt.Println("主键查询", id)
		// 通过主键查行记录
		query := `select id, kind, nickname from nest_user.profile where id=?`
		return conn.Query(&profile, query, id)
	})
	if err == nil {
		fmt.Println(profile)
	} else if err == ErrNotFound {
		fmt.Println("查无此人")
	} else {
		fmt.Println("操作出错：", err)
	}

	fmt.Println("Total", atomic.LoadUint64(&cacheStat.Total))
	fmt.Println("Hit", atomic.LoadUint64(&cacheStat.Hit))
	fmt.Println("Miss", atomic.LoadUint64(&cacheStat.Miss))
	fmt.Println("DbFails", atomic.LoadUint64(&cacheStat.DbFails))
}

func TestCachedConn_Exec(t *testing.T) {
	//userIdKey := fmt.Sprintf("%s%v", cacheUserIdPrefix)
}

func TestCachedConn_GetCache(t *testing.T) {
	resetStats()
	r := redis.NewRedis("192.168.0.166:6800", redis.StandaloneMode)
	conn := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true")
	c := NewCachedConn(conn, r, cache.WithExpires(10*time.Second))

	var v string
	var err error

	// 未写先读
	err = c.GetCache("ping", &v)
	assert.Equal(t, ErrNotFound, err)
	fmt.Println(v)

	// 写完再读
	err = c.SetCache("ping", "pong")
	err = c.GetCache("ping", &v)
	assert.Nil(t, err)
	assert.Equal(t, "pong", v)
	fmt.Println(v)
}

func TestCachedConn_Stat(t *testing.T) {
	resetStats()
	r := redis.NewRedis("192.168.0.166:6800", redis.StandaloneMode)
	conn := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true")
	c := NewCachedConn(conn, r, cache.WithExpires(10*time.Second))

	var err error
	for i := 0; i < 10; i++ {
		var str string
		err = c.Query(&str, "sqlx/name", func(conn Conn, v interface{}) error {
			//*v.(*string) = "goa"
			*v.(*string) = "哈哈"
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		fmt.Println(str)
	}

	fmt.Println(atomic.LoadUint64(&cacheStat.Total))
	fmt.Println(atomic.LoadUint64(&cacheStat.Hit))
	fmt.Println(atomic.LoadUint64(&cacheStat.Miss))
	fmt.Println(atomic.LoadUint64(&cacheStat.DbFails))

	var str string
	fmt.Println(c.GetCache("sqlx/name", &str))
	fmt.Println(str)
}

func TestCachedConn_QueryIndex_NoCache(t *testing.T) {
	resetStats()
	r := redis.NewRedis("192.168.0.166:6800", redis.StandaloneMode)
	conn := NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_user?parseTime=true")
	c := NewCachedConn(conn, r, cache.WithExpires(10*time.Second))

	//var err error
	var str string
	c.QueryIndex(&str, "sqlx/index", func(primaryKey interface{}) string {
		// 返回主键的缓存键
		return fmt.Sprintf("%s/1234", primaryKey)
	}, func(conn Conn, v interface{}) (interface{}, error) {
		// 根据索引键查主键
		fmt.Println("")
		fmt.Println(v)
		*v.(*string) = "我爱苏州"
		return "primary", nil
	}, func(conn Conn, v, primaryKey interface{}) error {
		return nil
	})
}

func resetStats() {
	atomic.StoreUint64(&cacheStat.Total, 0)
	atomic.StoreUint64(&cacheStat.Hit, 0)
	atomic.StoreUint64(&cacheStat.Miss, 0)
	atomic.StoreUint64(&cacheStat.DbFails, 0)
}
