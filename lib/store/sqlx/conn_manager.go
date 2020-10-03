package sqlx

import (
	"database/sql"
	"goa/lib"
	"io"
	"sync"
	"time"
)

const (
	maxOpenConns = 64          // 允许最大的打开连接数
	maxIdleConns = 64          // 允许的最大空闲连接数
	maxLifetime  = time.Minute // 允许的最大连接空闲时间
)

var connManager = lib.NewResourceManager()

type cachedConn struct {
	*sql.DB
	once sync.Once
}

// getConn 从缓存池中获取可复用的数据库
// TODO 待考虑 Conn 自动故障迁移
func getConn(driverName, dataSourceName string) (*sql.DB, error) {
	conn, err := getCachedConn(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// 尝试连接数据库（仅在第一次调用 getConn 时，下述方法才调用）
	conn.once.Do(func() {
		err = conn.Ping()
	})
	if err != nil {
		return nil, err
	}

	return conn.DB, nil
}

// getCachedConn 从缓存池中获取连接
func getCachedConn(driverName, dataSourceName string) (*cachedConn, error) {
	cc, err := connManager.Get(dataSourceName, func() (io.Closer, error) {
		conn, err := newConn(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return &cachedConn{DB: conn}, nil
	})
	if err != nil {
		return nil, err
	}
	return cc.(*cachedConn), nil
}

// newConn 新建数据库连接
func newConn(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(maxLifetime)

	return db, nil
}
