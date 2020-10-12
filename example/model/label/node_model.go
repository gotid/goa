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
	nodeFieldNames            = builder.FieldNames(&Node{})
	nodeFields                = strings.Join(nodeFieldNames, ",")
	nodeFieldsAutoSet         = strings.Join(stringx.Remove(nodeFieldNames, "id", "created_at", "updated_at"), ",")
	nodeFieldsWithPlaceHolder = strings.Join(stringx.Remove(nodeFieldNames, "id", "created_at", "updated_at"), "=?,") + "=?"

	cacheNodeIdPrefix = "cache#node#id#"
)

type (
	NodeModel struct {
		sqlx.CachedConn
		table string
	}

	Node struct {
		Id        int64     `db:"id"`
		Name      string    `db:"name"`       // 后台标签
		ParentId  int64     `db:"parent_id"`  // 父级编号
		ParentIds string    `db:"parent_ids"` // 父级路径
		Level     int64     `db:"level"`      // 节点级别
		IsValid   int64     `db:"is_valid"`   // 是否有效
		Sort      int64     `db:"sort"`       // 同级排序
		Txt       string    `db:"txt"`        // 描述文本
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
)

func NewNodeModel(conn sqlx.Conn, clusterConf cache.ClusterConf, table string) *NodeModel {
	return &NodeModel{
		CachedConn: sqlx.NewCachedConnWithCluster(conn, clusterConf),
		table:      table,
	}
}

func (m *NodeModel) Insert(data Node) (sql.Result, error) {
	query := `insert into ` + m.table + ` (` + nodeFieldsAutoSet + `) values (?, ?, ?, ?, ?, ?, ?)`
	return m.ExecNoCache(query, data.Name, data.ParentId, data.ParentIds, data.Level, data.IsValid, data.Sort, data.Txt)
}

func (m *NodeModel) FindOne(id int64) (*Node, error) {
	nodeIdKey := fmt.Sprintf("%s%v", cacheNodeIdPrefix, id)
	var dest Node
	err := m.Query(&dest, nodeIdKey, func(conn sqlx.Conn, v interface{}) error {
		query := `select ` + nodeFields + ` from ` + m.table + ` where id = ? limit 1`
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

func (m *NodeModel) Update(data Node) error {
	nodeIdKey := fmt.Sprintf("%s%v", cacheNodeIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + nodeFieldsWithPlaceHolder + ` where id = ?`
		return conn.Exec(query, data.Name, data.ParentId, data.ParentIds, data.Level, data.IsValid, data.Sort, data.Txt, data.Id)
	}, nodeIdKey)
	return err
}

func (m *NodeModel) Delete(id int64) error {

	nodeIdKey := fmt.Sprintf("%s%v", cacheNodeIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `delete from ` + m.table + ` where id = ?`
		return conn.Exec(query, id)
	}, nodeIdKey)
	return err
}
