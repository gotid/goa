package gen

import (
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
)

func genTag(fieldName string) (string, error) {
	if fieldName == "" {
		return fieldName, nil
	}

	output, err := util.With("tag").Parse(tpl.Tag).Execute(map[string]interface{}{
		"field": fieldName,
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
