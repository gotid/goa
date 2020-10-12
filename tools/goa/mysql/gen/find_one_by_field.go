package gen

import (
	"fmt"
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
	"strings"
)

func genFindOneByField(table Table, withCache bool) (string, error) {
	t := util.With("findOneByField").Parse(tpl.FindOneByField)
	var list []string
	upperTable := table.Name.ToCamel()
	for _, field := range table.Fields {
		if field.IsPrimaryKey || !field.IsUniqueKey {
			continue
		}
		upperField := field.Name.ToCamel()
		output, err := t.Execute(map[string]interface{}{
			"upperTable":                upperTable,
			"upperField":                upperField,
			"in":                        fmt.Sprintf("%s %s", stringx.From(upperField).UnTitle(), field.DataType),
			"withCache":                 withCache,
			"cacheKeyName":              table.CacheKeys[field.Name.Source()].KeyName,
			"cacheKeyExpression":        table.CacheKeys[field.Name.Source()].KeyExpression,
			"primaryKeyLeft":            table.CacheKeys[table.PrimaryKey.Name.Source()].Left,
			"lowerTable":                stringx.From(upperTable).UnTitle(),
			"lowerField":                stringx.From(upperField).UnTitle(),
			"upperStartCamelPrimaryKey": table.PrimaryKey.Name.ToCamel(),
			"originalField":             field.Name.Source(),
			"originalPrimaryField":      table.PrimaryKey.Name.Source(),
		})
		if err != nil {
			return "", err
		}
		list = append(list, output.String())
	}
	return strings.Join(list, "\n"), nil
}
