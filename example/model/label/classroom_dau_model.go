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
	classroomDauFieldNames            = builder.FieldNames(&ClassroomDau{})
	classroomDauFields                = strings.Join(classroomDauFieldNames, ",")
	classroomDauFieldsAutoSet         = strings.Join(stringx.Remove(classroomDauFieldNames, "id", "created_at", "updated_at"), ",")
	classroomDauFieldsWithPlaceHolder = strings.Join(stringx.Remove(classroomDauFieldNames, "id", "created_at", "updated_at"), "=?,") + "=?"

	cacheClassroomDauIdPrefix        = "cache#classroomDau#id#"
	cacheClassroomDauClassroomPrefix = "cache#classroomDau#classroom#"
)

type (
	ClassroomDauModel struct {
		sqlx.CachedConn
		table string
	}

	ClassroomDau struct {
		Id         int64     `db:"id"`
		Classroom  string    `db:"classroom"`
		User       string    `db:"user"`
		Count      int64     `db:"count"`
		IsOvertime time.Time `db:"is_overtime"`
	}
)

func NewClassroomDauModel(conn sqlx.Conn, clusterConf cache.ClusterConf, table string) *ClassroomDauModel {
	return &ClassroomDauModel{
		CachedConn: sqlx.NewCachedConnWithCluster(conn, clusterConf),
		table:      table,
	}
}

func (m *ClassroomDauModel) Insert(data ClassroomDau) (sql.Result, error) {
	query := `insert into ` + m.table + ` (` + classroomDauFieldsAutoSet + `) values (?, ?, ?, ?)`
	return m.ExecNoCache(query, data.Classroom, data.User, data.Count, data.IsOvertime)
}

func (m *ClassroomDauModel) FindOne(id int64) (*ClassroomDau, error) {
	classroomDauIdKey := fmt.Sprintf("%s%v", cacheClassroomDauIdPrefix, id)
	var dest ClassroomDau
	err := m.Query(&dest, classroomDauIdKey, func(conn sqlx.Conn, v interface{}) error {
		query := `select ` + classroomDauFields + ` from ` + m.table + ` where id = ? limit 1`
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

func (m *ClassroomDauModel) FindOneByClassroom(classroom string) (*ClassroomDau, error) {
	classroomDauClassroomKey := fmt.Sprintf("%s%v", cacheClassroomDauClassroomPrefix, classroom)
	var dest ClassroomDau
	err := m.QueryIndex(&dest, classroomDauClassroomKey, func(primary interface{}) string {
		// 主键的缓存键
		return fmt.Sprintf("%s%v", cacheClassroomDauIdPrefix, primary)
	}, func(conn sqlx.Conn, v interface{}) (i interface{}, e error) {
		// 无索引建——主键对应缓存，通过索引键查目标行
		query := `select ` + classroomDauFields + ` from ` + m.table + ` where classroom = ? limit 1`
		if err := conn.Query(&dest, query, classroom); err != nil {
			return nil, err
		}
		return dest.Id, nil
	}, func(conn sqlx.Conn, v, primary interface{}) error {
		// 如果有索引建——主键对应缓存，则通过主键直接查目标航
		query := `select ` + classroomDauFields + ` from ` + m.table + ` where id = ? limit 1`
		return conn.Query(v, query, primary)
	})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}
}

func (m *ClassroomDauModel) Update(data ClassroomDau) error {
	classroomDauIdKey := fmt.Sprintf("%s%v", cacheClassroomDauIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + classroomDauFieldsWithPlaceHolder + ` where id = ?`
		return conn.Exec(query, data.Classroom, data.User, data.Count, data.IsOvertime, data.Id)
	}, classroomDauIdKey)
	return err
}

func (m *ClassroomDauModel) Delete(id int64) error {
	data, err := m.FindOne(id)
	if err != nil {
		return err
	}

	classroomDauIdKey := fmt.Sprintf("%s%v", cacheClassroomDauIdPrefix, id)
	classroomDauClassroomKey := fmt.Sprintf("%s%v", cacheClassroomDauClassroomPrefix, data.Classroom)
	_, err = m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `delete from ` + m.table + ` where id = ?`
		return conn.Exec(query, id)
	}, classroomDauIdKey, classroomDauClassroomKey)
	return err
}
