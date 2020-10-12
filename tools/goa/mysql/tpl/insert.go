package tpl

var Insert = `
func (m *{{.upperTable}}Model) Insert(data {{.upperTable}}) (sql.Result, error) {
	query := ` + "`" + `insert into ` + "`" + ` + m.table + ` + "` (` + " + `{{.lowerTable}}FieldsAutoSet` + " + `) values ({{.args}})` " + `
	return m.{{if .withCache}}ExecNoCache{{else}}conn.Exec{{end}}(query, {{.values}})
}
`
