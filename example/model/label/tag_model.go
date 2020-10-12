package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/z-sdk/goa/lib/store/cache"
	"github.com/z-sdk/goa/lib/store/sqlx"
	"github.com/z-sdk/goa/lib/stringx"
	"github.com/z-sdk/goa/tools/goa/mysql/builder"
)

var (
	tagFieldNames            = builder.FieldNames(&Tag{})
	tagFields                = strings.Join(tagFieldNames, ",")
	tagFieldsAutoSet         = strings.Join(stringx.Remove(tagFieldNames, "id", "created_at", "updated_at"), ",")
	tagFieldsWithPlaceHolder = strings.Join(stringx.Remove(tagFieldNames, "id", "created_at", "updated_at"), "=?,") + "=?"

	cacheTagIdPrefix = "cache#tag#id#"
)

type (
	TagModel struct {
		sqlx.CachedConn
		table string
	}

	Tag struct {
		Id        int64     `db:"id"`
		Name      string    `db:"name"`       // 前台标签
		NodeIds   string    `db:"node_ids"`   // 后台标签聚合关系
		NodeNames string    `db:"node_names"` // 后台标签的名称 JSON格式
		IsValid   int64     `db:"is_valid"`   // 状态，1启用，0禁用
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
		Sort      int64     `db:"sort"` // 同级排序
	}
)

func NewTagModel(conn sqlx.Conn, clusterConf cache.ClusterConf, table string) *TagModel {
	return &TagModel{
		CachedConn: sqlx.NewCachedConnWithCluster(conn, clusterConf),
		table:      table,
	}
}

func (m *TagModel) Insert(data Tag) (sql.Result, error) {
	query := `insert into ` + m.table + ` (` + tagFieldsAutoSet + `) values (?, ?, ?, ?, ?)`
	return m.ExecNoCache(query, data.Name, data.NodeIds, data.NodeNames, data.IsValid, data.Sort)
}

func (m *TagModel) FindOne(id int64) (*Tag, error) {
	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, id)
	var dest Tag
	err := m.Query(&dest, tagIdKey, func(conn sqlx.Conn, v interface{}) error {
		query := `select ` + tagFields + ` from ` + m.table + ` where id = ? limit 1`
		return conn.Query(v, query, id)
	})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}
}

func (m *TagModel) Update(data Tag) error {
	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + tagFieldsWithPlaceHolder + ` where id = ?`
		return conn.Exec(query, data.Name, data.NodeIds, data.NodeNames, data.IsValid, data.Sort, data.Id)
	}, tagIdKey)
	return err
}

func (m *TagModel) Delete(id int64) error {

	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `delete from ` + m.table + ` where id = ?`
		return conn.Exec(query, id)
	}, tagIdKey)
	return err
}
