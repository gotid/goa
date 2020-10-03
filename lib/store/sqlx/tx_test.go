package sqlx

import (
	"testing"
)

func TestTxSession_Exec(t *testing.T) {
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)

	err := db.Transact(func(session Session) error {
		session.Exec("insert into book(id, book) values(5, '3rsf')")
		//a := 1
		//b := 0
		//fmt.Println(a / b)
		session.Exec("insert into book(id, book) values(6, 'sadsf')")
		//if err != nil {
		//	return err
		//}
		//insertId, err := result.LastInsertId()
		//if err != nil {
		//	return err
		//}
		//fmt.Println("新增编号", insertId)
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
