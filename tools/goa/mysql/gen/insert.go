package gen

import (
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
	"strings"
)

func genInsert(table Table, withCache bool) (string, error) {
	args := make([]string, 0)
	values := make([]string, 0)

	for _, field := range table.Fields {
		camelField := field.Name.ToCamel()
		if camelField == "CreatedAt" || camelField == "UpdatedAt" {
			continue
		}
		if field.IsPrimaryKey && table.PrimaryKey.AutoIncrement {
			continue
		}

		args = append(args, "?")
		values = append(values, "data."+camelField)
	}
	upperTable := table.Name.ToCamel()
	output, err := util.With("insert").Parse(tpl.Insert).Execute(map[string]interface{}{
		"withCache":  withCache,
		"upperTable": upperTable,
		"lowerTable": stringx.From(upperTable).UnTitle(),
		"args":       strings.Join(args, ", "),
		"values":     strings.Join(values, ", "),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
