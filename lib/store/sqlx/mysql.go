package sqlx

import (
	"github.com/go-sql-driver/mysql"
	//_ "github.com/go-command-driver/mysql"
)

const (
	ErrDuplicateEntryCode uint16 = 1062
)

// NewMySQL 创建 MySQL 数据库实例
func NewMySQL(dataSourceName string, opts ...Option) Conn {
	opts = append(opts, withMySQLAcceptable())
	return NewConn("mysql", dataSourceName, opts...)
}

func withMySQLAcceptable() Option {
	return func(c *conn) {
		c.accept = mysqlAcceptable
	}
}

func mysqlAcceptable(reqError error) bool {
	if reqError == nil {
		return true
	}

	sqlError, ok := reqError.(*mysql.MySQLError)
	if !ok {
		return false
	}

	switch sqlError.Number {
	case ErrDuplicateEntryCode:
		return true
	default:
		return false
	}
}
