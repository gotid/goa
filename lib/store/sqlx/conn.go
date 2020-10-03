package sqlx

import (
	"database/sql"
	"errors"
	"time"
)

const (
	// tagFieldKey 结构体字段中标记的数据库字段键名
	tagFieldKey = "db"

	// 慢日志阈值
	slowThreshold = 500 * time.Millisecond
)

var (
	ErrNotFound             = errors.New("没有结果集")
	ErrNotSettable          = errors.New("扫描目标不可设置")
	ErrUnsupportedValueType = errors.New("不支持的扫描目标类型")
	ErrNotReadableValue     = errors.New("无法读取的值，检查结构体字段是否为大写开头")
)

type (
	// StmtConn 语句执行和查询接口
	//StmtConn interface {
	//	Query(dest interface{}, args ...interface{}) error
	//	Exec(args ...interface{}) (sql.Result, error)
	//	Close() error
	//}

	// Session 提供基于 SQL 执行和查询
	Session interface {
		Query(dest interface{}, query string, args ...interface{}) error
		Exec(query string, args ...interface{}) (sql.Result, error)
		//Prepare(query string) (StmtConn, error)
	}

	// Conn 封装数据库会话和事务操作
	Conn interface {
		Session
		Transact(fn func(session Session) error) error
	}

	sessionConn interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		Exec(query string, args ...interface{}) (sql.Result, error)
	}

	// stmtConn 预编译连接
	//stmtConn interface {
	//	Query(args ...interface{}) (*sql.Rows, error)
	//	Exec(args ...interface{}) (sql.Result, error)
	//}

	// Option 是一个可选的数据库增强函数
	Option func(ins *conn)

	// conn 数据库实例，封装SQL和参数为语句、开启事务、支持断路器保护
	conn struct {
		driverName     string    // 驱动名称，支持 mysql/postgres/clickhouse
		dataSourceName string    // 数据源名称 Data Source Name，既数据库连接字符串
		beginTx        beginTxFn // 可开始事务
	}

	// statement 预编译语句会话：将查询封装为预编译语句，供底层查询和执行
	//statement struct {
	//	stmt *sql.Stmt
	//}
)

// NewConn 创建指定驱动和数据源地址的 Conn 实例
func NewConn(driverName, dataSourceName string, opts ...Option) Conn {
	prefectDSN(&dataSourceName)

	c := &conn{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		beginTx:        beginTx,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ----------------- conn 实现方法 ↓ ----------------- //

// Query 执行数据库查询并将结果扫描至结果。
// 如果 dest 字段不写tag的话，系统按顺序配对，此时需要与sql中的查询字段顺序一致
// 如果 dest 字段写了tag的话，系统按名称配对，此时可以和sql中的查询字段顺序不同
func (c *conn) Query(dest interface{}, query string, args ...interface{}) error {
	db, err := getConn(c.driverName, c.dataSourceName)
	if err != nil {
		logConnError(c.dataSourceName, err)
		return err
	}
	return doQuery(db, func(rows *sql.Rows) error {
		return scan(rows, dest)
	}, query, args...)
}

func (c *conn) Exec(query string, args ...interface{}) (sql.Result, error) {
	db, err := getConn(c.driverName, c.dataSourceName)
	if err != nil {
		logConnError(c.dataSourceName, err)
		return nil, err
	}
	return doExec(db, query, args...)
}

func (c *conn) Transact(fn func(session Session) error) error {
	return doTransaction(c, c.beginTx, fn)
}

// Prepare 创建一个稍后查询或执行的预编译语句
//func (c *conn) Prepare(query string) (stmt StmtConn, err error) {
//	db, err := getConn(c.driverName, c.dataSourceName)
//	if err != nil {
//		logConnError(c.dataSourceName, err)
//		return nil, err
//	}
//	if st, err := db.Prepare(query); err != nil {
//		return nil, err
//	} else {
//		stmt = statement{stmt: st}
//		return stmt, nil
//	}
//}

//func (s statement) Query(dest interface{}, args ...interface{}) error {
//	panic("implement me")
//}
//
//func (s statement) Exec(args ...interface{}) (sql.Result, error) {
//	panic("implement me")
//}
//
//func (s statement) Close() error {
//	panic("implement me")
//}
