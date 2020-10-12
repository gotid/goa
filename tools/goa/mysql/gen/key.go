package gen

import (
	"fmt"
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/parser"
)

type Key struct {
	Pattern string // cacheUserIdPrefix = "cache#user#id#"
	Left    string // cacheUserIdPrefix
	Right   string // cache#user#id#

	KeyName           string // userIdKey
	KeyExpression     string // userIdKey := fmt.Sprintf("cache#user#id#%v", userId)
	DataKeyExpression string // userIdKey := fmt.Sprintf("cache#user#id#%v", data.userId)
	RespKeyExpression string // userIdKey := fmt.Sprintf("cache#user#id#%v", resp.userId)
}

// 根据表字段中的唯一索引建和主键，自动生成缓存键相关代码
func genCacheKeys(table parser.Table) (map[string]Key, error) {
	fields := table.Fields
	m := make(map[string]Key)
	camelTableName := table.Name.ToCamel()
	lowerStartCamelTableName := stringx.From(camelTableName).UnTitle()
	for _, field := range fields {
		if !field.IsUniqueKey && !field.IsPrimaryKey {
			continue
		}

		camelFieldName := field.Name.ToCamel()
		lowerStartCamelFieldName := stringx.From(camelFieldName).UnTitle()
		left := fmt.Sprintf("cache%s%sPrefix", camelTableName, camelFieldName)
		right := fmt.Sprintf("cache#%s#%s#", lowerStartCamelTableName, lowerStartCamelFieldName)
		keyName := fmt.Sprintf("%s%sKey", lowerStartCamelTableName, camelFieldName)
		m[field.Name.Source()] = Key{
			Pattern:           fmt.Sprintf(`%s = "%s"`, left, right),
			Left:              left,
			Right:             right,
			KeyName:           keyName,
			KeyExpression:     fmt.Sprintf(`%s := fmt.Sprintf("%s%s", %s, %s)`, keyName, "%s", "%v", left, lowerStartCamelFieldName),
			DataKeyExpression: fmt.Sprintf(`%s := fmt.Sprintf("%s%s", %s, data.%s)`, keyName, "%s", "%v", left, camelFieldName),
			RespKeyExpression: fmt.Sprintf(`%s := fmt.Sprintf("%s%s", %s, resp.%s)`, keyName, "%s", "%v", left, camelFieldName),
		}
	}
	return m, nil
}
