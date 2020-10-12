package gen

import (
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
	"strings"
)

func genUpdate(table Table, withCache bool) (string, error) {
	values := make([]string, 0)
	for _, field := range table.Fields {
		upperField := field.Name.ToCamel()
		if upperField == "CreatedAt" || upperField == "UpdatedAt" || field.IsPrimaryKey {
			continue
		}
		values = append(values, "data."+upperField)
	}

	values = append(values, "data."+table.PrimaryKey.Name.ToCamel())
	upperTable := table.Name.ToCamel()
	output, err := util.With("update").
		Parse(tpl.Update).
		Execute(map[string]interface{}{
			"withCache":          withCache,
			"upperTable":         upperTable,
			"primaryCacheKey":    table.CacheKeys[table.PrimaryKey.Name.Source()].DataKeyExpression,
			"primaryKeyName":     table.CacheKeys[table.PrimaryKey.Name.Source()].KeyName,
			"lowerTable":         stringx.From(upperTable).UnTitle(),
			"originalPrimaryKey": table.PrimaryKey.Name.Source(),
			"values":             strings.Join(values, ", "),
		})
	if err != nil {
		return "", nil
	}
	return output.String(), nil
}
