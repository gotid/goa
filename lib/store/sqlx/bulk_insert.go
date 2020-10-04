package sqlx

import "database/sql"

const (
	// 最大批量插入数量
	maxBulkRows = 1000
)

type (
	// 批量插入器结构
	BulkInserter struct {
		stmt     bulkStmt
		inserter *dbInserter
	}

	// 批量执行语句结构
	bulkStmt struct {
		query  string
		prefix string
		suffix string
	}

	// 数据库插入器结构
	dbInserter struct {
		conn          Conn
		stmt          bulkStmt
		querys        []string
		resultHandler ResultHandler
	}

	// 执行结果处理器
	ResultHandler func(sql.Result, error)
)

func (in *dbInserter) AddTask(task interface{}) bool {
	in.querys = append(in.querys, task.(string))
	return len(in.querys) >= maxBulkRows
}
