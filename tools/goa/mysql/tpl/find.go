package tpl

// 通过id查询
var FindOne = `
func (m *{{.upperTable}}Model) FindOne({{.primaryKey}} {{.dataType}}) (*{{.upperTable}}, error) {
	{{if .withCache}}{{.cacheKeyExpression}}
	var dest {{.upperTable}}
	err := m.Query(&dest, {{.cacheKeyName}}, func(conn sqlx.Conn, v interface{}) error {
		query := ` + "`" + `select ` + "`" + ` + {{.lowerTable}}Fields + ` + "`" + ` from ` + "` + " + `m.table ` + " + `" + ` where {{.originalPrimaryKey}} = ? limit 1` + "`" + `
		return conn.Query(v, query, {{.primaryKey}})
	})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}{{else}}query := ` + "`" + `select ` + "`" + ` + {{.lowerTable}}Fields + ` + "`" + ` from ` + "` + " + `m.table ` + " + `" + ` where {{.originalPrimaryKey}} = ? limit 1` + "`" + `
	var dest {{.upperTable}}
	err := m.conn.Query(&dest, query, {{.primaryKey}})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}{{end}}
}
`

// 通过指定字段查询
var FindOneByField = `
func (m *{{.upperTable}}Model) FindOneBy{{.upperField}}({{.in}}) (*{{.upperTable}}, error) {
	{{if .withCache}}{{.cacheKeyExpression}}
	var dest {{.upperTable}}
	err := m.QueryIndex(&dest, {{.cacheKeyName}}, func(primary interface{}) string {
		// 主键的缓存键
		return fmt.Sprintf("%s%v", {{.primaryKeyLeft}}, primary)
	}, func(conn sqlx.Conn, v interface{}) (i interface{}, e error) {
		// 无索引建——主键对应缓存，通过索引键查目标行
		query := ` + "`" + `select ` + "`" + ` + {{.lowerTable}}Fields + ` + "`" + ` from ` + "` + " + `m.table ` + " + `" + ` where {{.originalField}} = ? limit 1` + "`" + `
		if err := conn.Query(&dest, query, {{.lowerField}}); err != nil {
			return nil, err
		}
		return dest.{{.upperStartCamelPrimaryKey}}, nil
	}, func(conn sqlx.Conn, v, primary interface{}) error {
		// 如果有索引建——主键对应缓存，则通过主键直接查目标航
		query := ` + "`" + `select ` + "`" + ` + {{.lowerTable}}Fields + ` + "`" + ` from ` + "` + " + `m.table ` + " + `" + ` where {{.originalPrimaryField}} = ? limit 1` + "`" + `
		return conn.Query(v, query, primary)
	})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}
}{{else}}var dest {{.upperTable}}
	query := ` + "`" + `select ` + "`" + ` + {{.lowerTable}}Fields + ` + "`" + ` from ` + "` + " + `m.table ` + " + `" + ` where {{.originalField}} = ? limit 1` + "`" + `
	err := m.conn.Query(&dest, query, {{.lowerField}})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}
}{{end}}
`
