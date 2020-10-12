package tpl

var Vars = `
var (
	{{.table}}FieldNames          = builder.FieldNames(&{{.camelTable}}{})
	{{.table}}Fields                = strings.Join({{.table}}FieldNames, ",")
	{{.table}}FieldsAutoSet         = strings.Join(stringx.Remove({{.table}}FieldNames, {{if .autoIncrement}}"{{.primaryKey}}",{{end}} "created_at", "updated_at"), ",")
	{{.table}}FieldsWithPlaceHolder = strings.Join(stringx.Remove({{.table}}FieldNames, "{{.primaryKey}}", "created_at", "updated_at"), "=?,") + "=?"

	{{if .withCache}}{{.cacheKeys}}{{end}}
)
`
