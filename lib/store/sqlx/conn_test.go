package sqlx

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

type Dict struct {
	Id             int64     `conn:"id"`               // 字典表 | 公共库
	ParentId       int64     `conn:"parent_id"`        // 字典分类ID
	Name           string    `conn:"name"`             // 字典名称
	Type           string    `conn:"type"`             // 字典类型
	CreateTime     time.Time `conn:"create_time"`      // 创建时间
	CreateUserId   int64     `conn:"create_user_id"`   // 创建人id
	CreateUserName string    `conn:"create_user_name"` // 创建人姓名
	UpdateTime     time.Time `conn:"update_time"`      // 更新时间
	UpdateUserId   int64     `conn:"update_user_id"`   // 更新人id
	UpdateUserName string    `conn:"update_user_name"` // 更新者姓名
	DeleteFlag     int64     `conn:"delete_flag"`      // 删除标记: 0删除|1未删除
}

func TestABC(t *testing.T) {
	dataSourceName := "root:qxqgqzx2018@tcp(192.168.0.17:3306)/nest_public?parseTime=true"
	db := NewMySQL(dataSourceName)

	var dictList []*Dict
	err := db.Query(&dictList, "select id, name, create_time from dict limit 0, 5")
	if err != nil {
		panic(err)
	}

	for _, dict := range dictList {
		fmt.Println(dict.Id, dict.Name, dict.CreateTime)
	}
}

func TestDbInstance_QueryRows(t *testing.T) {
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)
	type AccountKinds []struct {
		Id   int
		Name string
	}

	var book struct {
		Name  string
		Total int
		Price float32
		kinds AccountKinds
	}
	type Books []struct {
		Total         int    `conn:"totalx"`
		Name          string `conn:"book"`
		NotExistField int    `conn:"y"`
	}

	var accountKinds AccountKinds
	var books Books
	var adminUsers []struct {
		Txt       string    `conn:"txt"`
		UserId    int       `conn:"user_id"`
		AdminId   int       `conn:"admin_id"`
		CreatedAt time.Time `conn:"created_at"`
	}

	// 查询测试
	errAccountKinds := db.Query(&accountKinds, "select id, value as name from nest_user.account_kind")
	errBook := db.Query(&book, "select book, count(0) total from book group by book order by total desc")
	errBooks := db.Query(&books, "select book, count(0) totalx, 1 as x, 2 as y from book group by book order by totalx desc")
	errAdminUsers := db.Query(&adminUsers, "select user_id, admin_id, txt, created_at from nest_admin.admin_user")

	if errAccountKinds != nil {
		t.Fatal(errAccountKinds)
	}

	if errBook != nil {
		t.Fatal(errBook)
	}

	if errBooks != nil {
		t.Fatal(errBooks)
	}

	book.kinds = accountKinds

	if errAdminUsers != nil {
		t.Fatal(errAdminUsers)
	}

	fmt.Println(book)

	for _, book := range books {
		fmt.Println(book)
	}

	for _, accountKind := range accountKinds {
		fmt.Println(accountKind)
	}

	for _, adminUser := range adminUsers {
		fmt.Println(adminUser)
	}
}

func TestDbInstance_Exec(t *testing.T) {
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)

	for i := 0; i < 100; i++ {
		_, _ = db.Exec("update nest_admin.admin_user set txt=? where id=?", "自在测试"+strconv.Itoa(i), i)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//lastInsertId, _ := result.LastInsertId()
		//rowsAffected, _ := result.RowsAffected()
		//fmt.Printf("LastInsertId: %d, RowsAffected: %d\n", lastInsertId, rowsAffected)
	}
}

func TestConnBreaker(t *testing.T) {
	//logx.Disable()
	//logx.SetLevel(logx.ErrorLevel)
	//dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true&timeout=10s&readTimeout=2s"
	//dataSourceName := "root:asdfasdf@tcp(218.244.143.31:3317)/nest_label?parseTime=true&timeout=1s&readTimeout=2s"
	dataSourceName := "root:asdfasdf@tcp(218.244.143.31:3317)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)
	var book struct {
		Book string `conn:"book"`
	}

	for i := 0; i < 100; i++ {
		_ = db.Query(&book, "select book from bookx limit ?", i)
	}
}
