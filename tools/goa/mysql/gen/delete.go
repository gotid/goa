package gen

import (
	"github.com/z-sdk/goa/lib/collection"
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
	"strings"
)

func genDelete(table Table, withCache bool) (string, error) {
	keySet := collection.NewSet()
	keyNamesSet := collection.NewSet()
	for fieldName, key := range table.CacheKeys {
		if fieldName == table.PrimaryKey.Name.Source() {
			keySet.AddStr(key.KeyExpression)
		} else {
			keySet.AddStr(key.DataKeyExpression)
		}
		keyNamesSet.AddStr(key.KeyName)
	}
	var containsIndexCache = false
	for _, item := range table.Fields {
		if item.IsUniqueKey {
			containsIndexCache = true
			break
		}
	}
	upperTable := table.Name.ToCamel()
	output, err := util.With("delete").Parse(tpl.Delete).Execute(map[string]interface{}{
		"upperStartCamelObject":     upperTable,
		"withCache":                 withCache,
		"containsIndexCache":        containsIndexCache,
		"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
		"dataType":                  table.PrimaryKey.DataType,
		"keys":                      strings.Join(keySet.KeysStr(), "\n"),
		"originalPrimaryKey":        table.PrimaryKey.Name.Source(),
		"keyValues":                 strings.Join(keyNamesSet.KeysStr(), ", "),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
