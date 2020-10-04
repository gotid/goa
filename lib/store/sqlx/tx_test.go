package sqlx

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestTxSession_Exec(t *testing.T) {
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)

	err := db.Transact(func(session Session) (err error) {

		// 执行1
		var result1 sql.Result
		result1, err = session.Exec("insert into book(book) values('3rsf')")

		// 故障点
		//a := 1
		//b := 0
		//fmt.Println(a / b)

		// 执行2
		_, err = session.Exec("insert into book(id, book) values(20, 'sadsf')")
		if err != nil {
			return err
		}

		var insertId int64
		insertId, err = result1.LastInsertId()
		if err != nil {
			return err
		}
		fmt.Println("新增编号", insertId)
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
