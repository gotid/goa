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

var dbManager = lib.NewResourceManager()

type cachedDB struct {
	*sql.DB
	once sync.Once
}

// getDB 从缓存池中获取可复用的数据库
// TODO 待考虑 DB 自动故障迁移
func getDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := getCachedDB(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// 尝试连接数据库（仅在第一次调用 getDB 时，下述方法才调用）
	db.once.Do(func() {
		err = db.Ping()
	})
	if err != nil {
		return nil, err
	}

	return db.DB, nil
}

// getCachedDB 从缓存池中获取数据库
func getCachedDB(driverName, dataSourceName string) (*cachedDB, error) {
	cdb, err := dbManager.Get(dataSourceName, func() (io.Closer, error) {
		db, err := newDB(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return &cachedDB{DB: db}, nil
	})
	if err != nil {
		return nil, err
	}
	return cdb.(*cachedDB), nil
}

// newDB 打开指定数据库驱动和数据源地址的数据库
func newDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(maxLifetime)

	return db, nil
}
