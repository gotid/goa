package sqlx

import _ "github.com/go-sql-driver/mysql"

// NewMySQL 创建 MySQL 数据库实例
func NewMySQL(dataSourceName string, opts ...Option) Conn {
	return NewConn("mysql", dataSourceName, opts...)
}
