package sqlx

import (
	"fmt"
	"testing"
	"time"
)

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
		Total         int    `db:"totalx"`
		Name          string `db:"book"`
		NotExistField int    `db:"y"`
	}

	var accountKinds AccountKinds
	var books Books
	var adminUsers []struct {
		Txt       string    `db:"txt"`
		UserId    int       `db:"user_id"`
		AdminId   int       `db:"admin_id"`
		CreatedAt time.Time `db:"created_at"`
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
	result, err := db.Exec("update nest_admin.admin_user set txt='自在测试1007' where id=?", 6)
	if err != nil {
		t.Fatal(err)
	}
	lastInsertId, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("LastInsertId: %d, RowsAffected: %d\n", lastInsertId, rowsAffected)
}

func TestConnBreaker(t *testing.T) {
	//logx.Disable()
	//logx.SetLevel(logx.ErrorLevel)
	//dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true&timeout=10s&readTimeout=2s"
	//dataSourceName := "root:asdfasdf@tcp(218.244.143.31:3317)/nest_label?parseTime=true&timeout=10s&readTimeout=2s"
	dataSourceName := "root:asdfasdf@tcp(218.244.143.31:3317)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)
	var book struct {
		Book string `db:"book"`
	}

	for i := 0; i < 10000; i++ {
		_ = db.Query(&book, "select book from book limit 1")
	}
}
