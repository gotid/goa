package tpl

var Types = `
type (
	{{.table}}Model struct {
		{{if .withCache}}sqlx.CachedConn{{else}}conn sqlx.Conn{{end}}
		table string
	}

	{{.table}} struct {
		{{.fields}}
	}
)
`
