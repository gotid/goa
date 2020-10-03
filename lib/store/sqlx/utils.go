package sqlx

import (
	"fmt"
	"goa/lib/mapping"
	"log"
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

// logSqlError TODO 日志开关控制
func logSqlError(sql string, err error) {
	if err != nil && err != ErrNotFound {
		log.Fatalf("[SQL] SQL: %s, 错误: %s", sql, err.Error())
	}
}
