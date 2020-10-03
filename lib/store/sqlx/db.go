package sqlx

import (
	"database/sql"
	"errors"
	"goa/lib/mapping"
	"log"
	"reflect"
	"strings"
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

// StmtSession 语句执行和查询接口
type StmtSession interface {
	Query(result interface{}, args ...interface{}) error
	Exec(args ...interface{}) (sql.Result, error)
	Close() error
}

// Session 提供基于 SQL 执行和查询
type Session interface {
	Prepare(query string) (StmtSession, error)
	Query(result interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// sessionConn TODO 封装事务查询和执行?
type TxSession interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// DB 封装数据库会话和事务操作
type DB interface {
	Session
	BeginTx(func(session Session) error) error
}

// dbInstance 数据库实例，封装SQL和参数为语句、开启事务、支持断路器保护
type dbInstance struct {
	driverName     string // 驱动名称，支持 mysql/postgres/clickhouse
	dataSourceName string // 数据源名称 Data Source Name，既数据库连接字符串
}

// preparedStmt 预编译语句：将查询封装为预编译语句，供底层查询和执行
type preparedStmt struct {
	stmt *sql.Stmt
}

// Option 是一个可选的数据库增强函数
type Option func(ins *dbInstance)

// NewDB 创建指定驱动和数据源地址的 DB 实例
func NewDB(driverName, dataSourceName string, opts ...Option) DB {
	prefectDSN(&dataSourceName)

	db := &dbInstance{
		driverName:     driverName,
		dataSourceName: dataSourceName,
	}
	for _, opt := range opts {
		opt(db)
	}
	return db
}

// 自动补全连接字符串
func prefectDSN(dataSourceName *string) {
	if strings.Count(*dataSourceName, "?") == 0 {
		*dataSourceName += "?"
	}
	var args []string

	if strings.Count(*dataSourceName, "parseTime=true") == 0 {
		args = append(args, "parseTime=true")
	}
	if strings.Count(*dataSourceName, "loc=Local") == 0 {
		args = append(args, "loc=Local")
	}
	*dataSourceName += strings.Join(args, "&")
}

// ----------------- dbInstance 实现方法 ↓ ----------------- //

// Prepare 创建一个用于稍后查询或执行的预编译语句
func (d *dbInstance) Prepare(query string) (stmt StmtSession, err error) {
	var db *sql.DB
	db, err = getDB(d.driverName, d.dataSourceName)
	if err != nil {
		// TODO 增加 logInstanceError
		return nil, err
	}
	if st, err := db.Prepare(query); err != nil {
		stmt = preparedStmt{stmt: st}
		return stmt, nil
	}
	return
}

// Query 执行数据库查询并将结果扫描至结果。
// 如果 result 字段不写tag的话，系统按顺序配对，此时需要与sql中的查询字段顺序一致
// 如果 result 字段写了tag的话，系统按名称配对，此时可以和sql中的查询字段顺序不同
func (d *dbInstance) Query(result interface{}, query string, args ...interface{}) error {
	db, err := getDB(d.driverName, d.dataSourceName)
	if err != nil {
		// 待处理数据库获取错误
		return err
	}
	return d.query(db, func(rows *sql.Rows) error {
		return scan(rows, result)
		//return unmarshalRows(result, rows, false)
	}, query, args...)
}

func (d *dbInstance) Exec(query string, args ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (d *dbInstance) BeginTx(f func(session Session) error) error {
	panic("implement me")
}

// query 执行数据库查询
func (d *dbInstance) query(db *sql.DB, scanner func(*sql.Rows) error, query string, args ...interface{}) error {
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
		//log.Printf("[SQL] 查询: %s", stmt)
	}

	if err != nil {
		logSqlError(stmt, err)
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

func scan(rows *sql.Rows, dest interface{}) error {
	// 验证接收目标必须为有效非空指针
	// TODO 为什么不直接验证 dest，而是反射的值？
	dv := reflect.ValueOf(dest)
	if err := mapping.ValidatePtr(&dv); err != nil {
		return err
	}

	// 将行数据扫描进目标结果
	dte := reflect.TypeOf(dest).Elem()
	dve := dv.Elem()
	switch dte.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		if dve.CanSet() {
			if !rows.Next() {
				if err := rows.Err(); err != nil {
					return err
				}
				return ErrNotFound
			}
			return rows.Scan(dest)
		} else {
			return ErrNotSettable
		}
	case reflect.Struct:
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return ErrNotFound
		}
		// 获取行的列名切片
		colNames, err := rows.Columns()
		if err != nil {
			return err
		}

		if values, err := mapStructFieldsToSlice(dve, colNames); err != nil {
			return err
		} else {
			return rows.Scan(values...)
		}
	case reflect.Slice:
		if !dve.CanSet() {
			return ErrNotSettable
		}

		ptr := dte.Elem().Kind() == reflect.Ptr
		appendFn := func(item reflect.Value) {
			if ptr {
				dve.Set(reflect.Append(dve, item))
			} else {
				dve.Set(reflect.Append(dve, reflect.Indirect(item)))
			}
		}
		fillFn := func(value interface{}) error {
			if dve.CanSet() {
				if err := rows.Scan(value); err != nil {
					return err
				} else {
					appendFn(reflect.ValueOf(value))
					return nil
				}
			}
			return ErrNotSettable
		}

		base := mapping.Deref(dte.Elem())
		switch base.Kind() {
		case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for rows.Next() {
				value := reflect.New(base)
				if err := fillFn(value.Interface()); err != nil {
					return err
				}
			}
		case reflect.Struct:
			// 获取行的列名切片
			colNames, err := rows.Columns()
			if err != nil {
				return err
			}

			for rows.Next() {
				value := reflect.New(base)
				if values, err := mapStructFieldsToSlice(value, colNames); err != nil {
					return err
				} else {
					if err := rows.Scan(values...); err != nil {
						return err
					} else {
						appendFn(value)
					}
				}
			}
		default:
			return ErrUnsupportedValueType
		}
		return nil
	default:
		return ErrUnsupportedValueType
	}
}

// 映射目标结构体字段到查询结果列，并赋初值
func mapStructFieldsToSlice(dve reflect.Value, columns []string) ([]interface{}, error) {
	columnValueMap, err := getColumnValueMap(dve)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	if len(columnValueMap) != 0 {
		for i, column := range columns {
			if value, ok := columnValueMap[column]; ok {
				values[i] = value
			} else {
				var anonymous interface{}
				values[i] = &anonymous
			}
		}
	} else {
		fields := getFields(dve)

		for i := 0; i < len(values); i++ {
			field := fields[i]
			switch field.Kind() {
			case reflect.Ptr:
				if !field.CanInterface() {
					return nil, ErrNotReadableValue
				}
				if field.IsNil() {
					baseValueType := mapping.Deref(field.Type())
					field.Set(reflect.New(baseValueType))
				}
				values[i] = field.Interface()
			default:
				if !field.CanAddr() || !field.Addr().CanInterface() {
					return nil, ErrNotReadableValue
				}
				values[i] = field.Addr().Interface()
			}
		}
	}

	return values, nil
}

// getColumnValueMap: 获取结构体字段中标记的列名——值映射关系
// 在编写字段tag的情况下，可以确保结构体字段和SQL选择列不一致的情况下不出错
func getColumnValueMap(dve reflect.Value) (map[string]interface{}, error) {
	t := mapping.Deref(dve.Type())
	size := t.NumField()
	result := make(map[string]interface{}, size)

	for i := 0; i < size; i++ {
		// 取字段标记中的列名，如`db:"total"` 中的 total
		columnName := getColumnName(t.Field(i))
		if len(columnName) == 0 {
			return nil, nil
		}

		// 读取指针字段或非指针字段的值
		field := reflect.Indirect(dve).Field(i)
		switch field.Kind() {
		case reflect.Ptr:
			if !field.CanInterface() {
				return nil, ErrNotReadableValue
			}
			if field.IsNil() {
				typ := mapping.Deref(field.Type())
				field.Set(reflect.New(typ))
			}
			result[columnName] = field.Interface()
		default:
			if !field.CanAddr() || !field.Addr().CanInterface() {
				return nil, ErrNotReadableValue
			}
			result[columnName] = field.Addr().Interface()
		}
	}

	return result, nil
}

// getColumnName 解析结构体字段中的数据库字段标记
func getColumnName(field reflect.StructField) string {
	tagName := field.Tag.Get(tagFieldKey)
	if len(tagName) == 0 {
		return ""
	} else {
		return strings.Split(tagName, ",")[0]
	}
}

// getFields 递归获取目标结构体的字段列表
func getFields(dve reflect.Value) []reflect.Value {
	var fields []reflect.Value
	v := reflect.Indirect(dve) // 指针取值

	// 递归目标字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// 指针取值
		if field.Kind() == reflect.Ptr && field.IsNil() {
			baseValueType := mapping.Deref(field.Type()) // 解引用，取值
			field.Set(reflect.New(baseValueType))
		}

		field = reflect.Indirect(field)
		structField := v.Type().Field(i)

		// 嵌套字段
		if field.Kind() == reflect.Struct && structField.Anonymous {
			fields = append(fields, getFields(field)...)
		} else {
			fields = append(fields, field)
		}
	}

	return fields
}

// ----------------- preparedStmt 实现方法 ↓ ----------------- //

func (s preparedStmt) Close() error {
	return s.stmt.Close()
}

func (s preparedStmt) Exec(args ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (s preparedStmt) Query(result interface{}, args ...interface{}) error {
	panic("implement me")
	// TODO 支持单行、单行字段，多行、多行字段
}
