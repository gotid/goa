package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"goa/lib/breaker"
	"time"
)

const (
	// 结构体字段中，数据库字段的标记名称
	tagName = "db"

	// 数据库慢日志阈值，用于记录慢查询和慢执行
	slowThreshold = 500 * time.Millisecond
)

var (
	ErrNotFound             = errors.New("没有结果集")
	ErrNotSettable          = errors.New("扫描目标不可设置")
	ErrUnsupportedValueType = errors.New("不支持的扫描目标类型")
	ErrNotReadableValue     = errors.New("无法读取的值，检查结构字段是否大写开头")
)

type (
	// Session 提供外部查询和执行的会话接口
	Session interface {
		Query(dest interface{}, query string, args ...interface{}) error
		Exec(query string, args ...interface{}) (sql.Result, error)
	}

	// 提供内部查询和执行的会话接口
	session interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		Exec(query string, args ...interface{}) (sql.Result, error)
	}

	// TransactFn 事务内部执行函数，传入事务会话
	TransactFn func(tx Session) error

	// Conn 提供外部数据库会话和事务的接口
	Conn interface {
		Session
		Transact(fn TransactFn) error
	}

	// conn 内部使用的数据库连接，封装查询、执行、事务及断路器支持
	conn struct {
		driverName     string          // 驱动名称，支持 mysql/postgres/clickhouse 等 sql-like
		dataSourceName string          // 数据源名称 Data Source Name，既数据库连接字符串
		beginTx        beginTxFn       // 可开始事务
		brk            breaker.Breaker // 断路器，用于后端故障拒绝服务
		accept         func(reqError error) bool
	}

	// Option 是一个可选的数据库增强函数
	Option func(c *conn)
)

// NewConn 新建指定数据库驱动和DSN的连接
func NewConn(driverName, dataSourceName string, opts ...Option) Conn {
	prefectDSN(&dataSourceName)

	c := &conn{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		beginTx:        beginTx,
		brk:            breaker.NewBreaker(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// 如果 dest 字段不写tag的话，系统按顺序配对，此时需要与sql中的查询字段顺序一致
// 如果 dest 字段写了tag的话，系统按名称配对，此时可以和sql中的查询字段顺序不同
func (c *conn) Query(dest interface{}, query string, args ...interface{}) error {
	var scanError error
	return c.brk.DoWithAcceptable(func() error {
		fmt.Println("获取连接并做数据库查询")
		fmt.Println()

		// 获取数据库连接
		db, err := getConn(c.driverName, c.dataSourceName)
		if err != nil {
			logConnError(c.dataSourceName, err)
			return err
		}

		// 做数据库查询
		return doQuery(db, func(rows *sql.Rows) error {
			scanError = scan(rows, dest)
			return scanError
		}, query, args...)
	}, func(reqError error) bool {
		return reqError == scanError || c.acceptable(reqError)
	})
}

func (c *conn) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	err = c.brk.DoWithAcceptable(func() error {
		//fmt.Println("获取连接并做数据库执行")
		//fmt.Println()

		// 获取数据库连接
		db, err := getConn(c.driverName, c.dataSourceName)
		if err != nil {
			logConnError(c.dataSourceName, err)
			return err
		}

		// 做数据库执行
		result, err = doExec(db, query, args...)
		return err
	}, c.acceptable)
	return
}

func (c *conn) Transact(fn TransactFn) error {
	return c.brk.DoWithAcceptable(func() error {
		fmt.Println("获取连接并做数据库事务")
		fmt.Println()
		return doTx(c, c.beginTx, fn)
	}, c.acceptable)
}

func (c *conn) acceptable(reqError error) bool {
	ok := reqError == nil ||
		reqError == sql.ErrNoRows ||
		reqError == sql.ErrTxDone
	if c.accept == nil {
		return ok
	} else {
		return ok || c.accept(reqError)
	}
}
