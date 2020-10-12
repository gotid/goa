package gen

import (
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
)

func genNew(table Table, withCache bool) (string, error) {
	output, err := util.With("new").Parse(tpl.New).Execute(map[string]interface{}{
		"withCache": withCache,
		"table":     table.Name.ToCamel(),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
