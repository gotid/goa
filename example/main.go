package main

import (
	"fmt"
	"github.com/z-sdk/goa/lib/store/sqlx"
	"time"
)

type (
	Node struct {
		Id        int64     `conn:"id"`
		Name      string    `conn:"name"`       // 后台标签
		ParentId  int64     `conn:"parent_id"`  // 父级编号
		ParentIds string    `conn:"parent_ids"` // 父级路径
		Level     int64     `conn:"level"`      // 节点级别
		IsValid   int64     `conn:"is_valid"`   // 是否有效
		Sort      int64     `conn:"sort"`       // 同级排序
		Txt       string    `conn:"txt"`        // 描述文本
		CreatedAt time.Time `conn:"created_at"`
		UpdatedAt time.Time `conn:"updated_at"`
	}
)

func FindNodes(parentIds string) ([]Node, error) {
	db := sqlx.NewMySQL("root:asdfasdf@tcp(192.168.0.166:3306)/nest_label")
	query := `select  id, name, created_at from node where parent_ids like ? limit 10`
	var resp []Node
	err := db.Query(&resp, query, parentIds)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, sqlx.ErrNotFound
	default:
		return nil, err
	}
}

func main() {
	nodes, err := FindNodes("/1/3/%")
	if err != nil {
		panic(err)
	}
	for i, node := range nodes {
		fmt.Println(i, node.Name, node.CreatedAt)
	}
}
