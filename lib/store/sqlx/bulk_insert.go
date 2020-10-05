package sqlx

import (
	"database/sql"
	"fmt"
	"goa/lib/dispatcher"
	"goa/lib/logx"
	"goa/lib/stringx"
	"strings"
	"time"
)

const (
	// 最大批量插入数量
	maxBulkRows = 1000

	// SQL 中的 values 标记
	valuesTag = "values"

	// 定期执行程序的的间隔执行时间
	flushInterval = time.Second
)

var emptyBulkStmt bulkStmt

type (
	// 批量插入器结构
	BulkInserter struct {
		// 批量插入语句
		stmt bulkStmt

		// 插入管理器
		manager *insertManager

		// 定时调度器
		dispatcher *dispatcher.PeriodicalDispatcher
	}

	// 批量插入语句结构
	// 形如 prefix valueFormat suffix
	bulkStmt struct {
		// values 之前的字符串
		prefix string

		// values 之后小括号内的值格式
		valueFormat string

		// valueFormat 值格式之后的剩余字符串
		suffix string
	}

	// 批量插入管理器
	insertManager struct {
		conn          Conn
		stmt          bulkStmt
		values        []string
		resultHandler ResultHandler
	}

	// 执行结果处理器
	ResultHandler func(sql.Result, error)
)

// NewBulkInserter 新建批量插入器
func NewBulkInserter(c Conn, stmt string) (*BulkInserter, error) {
	insertStmt, err := parseBulkInsertStmt(stmt)
	if err != nil {
		return nil, err
	}

	manager := &insertManager{
		conn: c,
		stmt: insertStmt,
	}

	return &BulkInserter{
		stmt:       insertStmt,
		manager:    manager,
		dispatcher: dispatcher.NewPeriodicalDispatcher(flushInterval, manager),
	}, nil
}

func (bi *BulkInserter) Insert(args ...interface{}) error {
	value, err := formatQuery(bi.stmt.valueFormat, args...)
	if err != nil {
		return err
	}

	bi.dispatcher.Add(value)

	return nil
}

func (bi *BulkInserter) Flush() {
	bi.dispatcher.Flush()
}

// SetResultHandler 设置结果处理器
func (bi *BulkInserter) SetRequestHandler(handler ResultHandler) {
	bi.dispatcher.Sync(func() {
		bi.manager.resultHandler = handler
	})
}

func (bi *BulkInserter) UpdateStmt(stmt string) error {
	newStmt, err := parseBulkInsertStmt(stmt)
	if err != nil {
		return err
	}

	bi.dispatcher.Flush()
	bi.dispatcher.Sync(func() {
		bi.manager.stmt = newStmt
	})

	return nil
}

func (bi *BulkInserter) UpdateOrDelete(fn func()) {
	bi.dispatcher.Flush()
	fn()
}

// --------------- 扩展 insertManager ↓ --------------- //

func (m *insertManager) Add(row interface{}) bool {
	m.values = append(m.values, row.(string))
	return len(m.values) >= maxBulkRows
}

func (m *insertManager) Execute(rows interface{}) {
	values := rows.([]string)
	if len(values) == 0 {
		return
	}

	stmtWithoutValues := m.stmt.prefix
	valuesStr := strings.Join(values, ",")
	stmt := strings.Join([]string{stmtWithoutValues, valuesStr}, " ")
	if len(m.stmt.suffix) > 0 {
		stmt = strings.Join([]string{stmt, m.stmt.suffix}, " ")
	}

	// 真正执行插入
	result, err := m.conn.Exec(stmt)

	// 处理执行结果
	if m.resultHandler != nil {
		m.resultHandler(result, err)
	} else if err != nil {
		logx.Errorf("[批量插入] SQL: %s, 错误: %s", stmt, err)
	}
}

func (m *insertManager) PopAll() interface{} {
	values := m.values
	m.values = nil
	return values
}

// --------------- 辅助方法 ↓ --------------- //
// parseBulkInsertStmt 解析批量插入语句
func parseBulkInsertStmt(stmt string) (bulkStmt, error) {
	// insert into users(id, name) values
	// (1, "张三")
	// (2, "李四")
	//
	// insert into users values
	// (1, "张三")
	// (2, "李四")

	lowerStmt := strings.ToLower(stmt)
	valuesPos := strings.Index(lowerStmt, valuesTag)
	if valuesPos <= 0 {
		return emptyBulkStmt, fmt.Errorf("sql 中没有找到 values 标记：%q", stmt)
	}

	// 尝试找出 values 之前定义的插入字段列数
	var numCols int
	// values 前面的右括号 insert into users(id, name【)】 values
	right := strings.LastIndexByte(lowerStmt[:valuesPos], ')')
	if right > 0 {
		// values insert into users【(】id, name) values
		left := strings.LastIndexByte(lowerStmt[:right], '(')
		if left > 0 {
			values := lowerStmt[left+1 : right]
			values = stringx.Filter(values, func(r rune) bool {
				// 去除sql语句中的空格和换行符
				return r == ' ' || r == '\t' || r == '\r' || r == '\n'
			})
			fields := strings.FieldsFunc(values, func(r rune) bool {
				return r == ','
			})
			numCols = len(fields)
		}
	}

	// 尝试找出 values 之后插入的值格式字符串
	var numArgs int
	var valueFormat string
	var suffix string
	left := strings.IndexByte(lowerStmt[valuesPos:], '(')
	if left > 0 {
		right = strings.IndexByte(lowerStmt[valuesPos+left:], ')')
		if right > 0 {
			values := lowerStmt[valuesPos+left : valuesPos+left+right]
			for _, x := range values {
				if x == '?' {
					numArgs++
				}
			}
			valueFormat = stmt[valuesPos+left : valuesPos+left+right+1]
			suffix = strings.TrimSpace(stmt[valuesPos+left+right+1:])
		}
	}

	if numArgs == 0 {
		return emptyBulkStmt, fmt.Errorf("没有变量占位符: %q", stmt)
	}
	if numCols > 0 && numCols != numArgs {
		return emptyBulkStmt, fmt.Errorf("列数和参数值个数不匹配: %q", stmt)
	}

	return bulkStmt{
		prefix:      stmt[:valuesPos+len(valuesTag)],
		valueFormat: valueFormat,
		suffix:      suffix,
	}, nil
}
