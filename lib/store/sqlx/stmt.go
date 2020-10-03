package sqlx

import (
	"database/sql"
	"log"
	"time"
)

func doQuery(db sessionConn, scanner func(*sql.Rows) error, query string, args ...interface{}) error {
	// 格式化后的查询字符串
	stmt, err := formatQuery(query, args...)
	if err != nil {
		return err
	}

	// 带有慢查询检测
	startTime := time.Now()
	rows, err := db.Query(query, args...)
	duration := time.Since(startTime)

	// TODO 日志开关控制 atomic
	if duration > slowThreshold {
		log.Printf("[SQL] 慢查询(%v) - %s", duration, stmt)
	} else {
		log.Printf("[SQL] 查询: %s", stmt)
	}

	if err != nil {
		logSqlError(stmt, err)
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

func doExec(db sessionConn, query string, args ...interface{}) (sql.Result, error) {
	// 格式化后的查询字符串
	stmt, err := formatQuery(query, args...)
	if err != nil {
		return nil, err
	}

	// 带有慢查询检测
	startTime := time.Now()
	result, err := db.Exec(query, args...)
	duration := time.Since(startTime)

	// TODO 日志开关控制 atomic
	if duration > slowThreshold {
		log.Printf("[SQL] 慢执行(%v) - %s", duration, stmt)
	} else {
		log.Printf("[SQL] 执行: %s", stmt)
	}

	if err != nil {
		logSqlError(stmt, err)
	}

	return result, err
}

//func doStmtQuery(conn stmtConn, scanner func(*sql.Rows) error, args ...interface{}) error {
//	// 格式化后的查询字符串
//	stmt := fmt.Sprint(args...)
//
//	// 带有慢查询检测
//	startTime := time.Now()
//	rows, err := conn.Query(args...)
//	duration := time.Since(startTime)
//
//	// TODO 日志开关控制 atomic
//	if duration > slowThreshold {
//		log.Printf("[SQL] 慢查询(%v) - %s", duration, stmt)
//	} else {
//		log.Printf("[SQL] 查询: %s", stmt)
//	}
//
//	if err != nil {
//		logSqlError(stmt, err)
//		return err
//	}
//	defer rows.Close()
//
//	return scanner(rows)
//}
