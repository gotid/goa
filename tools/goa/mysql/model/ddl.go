package model

import (
	"github.com/z-sdk/goa/lib/store/sqlx"
)

type (
	Model struct {
		conn sqlx.Conn
	}

	DDL struct {
		Table string `db:"Table"`
		DDL   string `db:"Create Table"`
	}
)

func NewModel(conn sqlx.Conn) *Model {
	return &Model{conn: conn}
}

func (m *Model) ShowDDL(tables ...string) ([]string, error) {
	var ddl []string
	for _, table := range tables {
		query := `show create table ` + table
		var resp DDL
		if err := m.conn.Query(&resp, query); err != nil {
			return nil, err
		}
		ddl = append(ddl, resp.DDL)
	}
	return ddl, nil
}
