package sqlx

import (
	"database/sql"
	"fmt"
	"goa/lib/logx"
	"goa/lib/mapping"
	"reflect"
	"strings"
)

// formatQuery 格式查询字符串和参数
func formatQuery(query string, args ...interface{}) (string, error) {
	argNum := len(args)
	if argNum == 0 {
		return query, nil
	}

	var b strings.Builder
	argIdx := 0
	for _, char := range query {
		if char != '?' {
			b.WriteRune(char)
		} else {
			if argIdx >= argNum {
				return "", fmt.Errorf("错误: 参数个数【少于】问号个数")
			}

			arg := args[argIdx]
			argIdx++

			switch at := arg.(type) {
			case bool:
				if at {
					b.WriteByte('1')
				} else {
					b.WriteByte('0')
				}
			case string:
				b.WriteByte('\'')
				b.WriteString(escape(at))
				b.WriteByte('\'')
			default:
				// 表示其他类型如 interface{} 的字符串形式
				b.WriteString(mapping.Repr(at))
			}
		}
	}

	if argIdx < argNum {
		return "", fmt.Errorf("参数个数【多于】问号个数")
	}

	return b.String(), nil
}

// escape 字符串转义
func escape(str string) string {
	var b strings.Builder

	for _, c := range str {
		switch c {
		case '\x00':
			b.WriteString(`\x00`)
		case '\r':
			b.WriteString(`\r`)
		case '\n':
			b.WriteString(`\n`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '"':
			b.WriteString(`\"`)
		case '\x1a':
			b.WriteString(`\x1a`)
		default:
			b.WriteRune(c)
		}
	}

	return b.String()
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
	if strings.HasSuffix(*dataSourceName, "?") {
		*dataSourceName += strings.Join(args, "&")
	} else {
		*dataSourceName += "&" + strings.Join(args, "&")
	}
}

func logSqlError(sql string, err error) {
	if err != nil && err != ErrNotFound {
		logx.Errorf("[SQL] %s >>> %s", err.Error(), sql)
	}
}

func logConnError(dsn string, err error) {
	dsn = desensitize(dsn)
	logx.Errorf("获取数据库实例失败 %s: %v", dsn, err)
}

// desensitize 脱敏：去除数据库连接中的账号密码
func desensitize(dsn string) string {
	pos := strings.LastIndex(dsn, "@")
	if pos >= 0 && pos+1 < len(dsn) {
		dsn = dsn[pos+1:]
	}
	return dsn
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
	tagName := field.Tag.Get(tagName)
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
