package tpl

var Model = `package model
{{.imports}}
{{.vars}}
{{.types}}
{{.new}}
{{.insert}}
{{.find}}
{{.update}}
{{.delete}}
`
