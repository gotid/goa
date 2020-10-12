package gen

import (
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
	"strings"
)

func genVars(table Table, withCache bool) (string, error) {
	keys := make([]string, 0)
	for _, key := range table.CacheKeys {
		keys = append(keys, key.Pattern)
	}
	camel := table.Name.ToCamel()
	output, err := util.With("var").
		Parse(tpl.Vars).
		GoFmt(true).
		Execute(map[string]interface{}{
			"table":         stringx.From(camel).UnTitle(),
			"camelTable":    camel,
			"cacheKeys":     strings.Join(keys, "\n"),
			"autoIncrement": table.PrimaryKey.AutoIncrement,
			"primaryKey":    table.PrimaryKey.Name.Source(),
			"withCache":     withCache,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
