package gen

import (
	"github.com/z-sdk/goa/tools/goa/mysql/tpl"
	"github.com/z-sdk/goa/tools/goa/util"
)

func genImports(withCache, timeImport bool) (string, error) {
	if withCache {
		buffer, err := util.With("import").Parse(tpl.Imports).Execute(map[string]interface{}{
			"time": timeImport,
		})
		if err != nil {
			return "", err
		}
		return buffer.String(), nil
	} else {
		buffer, err := util.With("import").Parse(tpl.ImportsNoCache).Execute(map[string]interface{}{
			"time": timeImport,
		})
		if err != nil {
			return "", err
		}
		return buffer.String(), nil
	}
}
