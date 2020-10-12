package converter

import (
	"fmt"
	"strings"
)

var (
	types = map[string]string{
		// 图省事，所有数据库整型字段都转为 int64
		"tinyint":    "int64",
		"smallint":   "int64",
		"mediumint":  "int64",
		"int":        "int64",
		"integer":    "int64",
		"bigint":     "int64",
		"float":      "float64",
		"double":     "float64",
		"decimal":    "float64",
		"date":       "time.Time",
		"time":       "string",
		"year":       "int64",
		"datetime":   "time.Time",
		"timestamp":  "time.Time",
		"char":       "string",
		"varchar":    "string",
		"tinyblob":   "string",
		"tinytext":   "string",
		"blob":       "string",
		"text":       "string",
		"mediumblob": "string",
		"mediumtext": "string",
		"longblob":   "string",
		"longtext":   "string",
	}
)

func ConvertDataType(dataBaseType string) (goDataType string, err error) {
	tp, ok := types[strings.ToLower(dataBaseType)]
	if !ok {
		err = fmt.Errorf("不识别的数据库类型: %s", dataBaseType)
		return
	}
	goDataType = tp
	return
}
