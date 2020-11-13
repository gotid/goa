package sqlx

import (
	"database/sql"
	"fmt"
	"strconv"
	"testing"
)

func TestTxSession_Exec(t *testing.T) {
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:33063)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)
	_ = db.Transact(addAccountCallback)
}

func addAccountCallback(tx Session) (err error) {
	for i := 0; i < 100; i++ {
		// 插入用户库——账号表
		var result sql.Result
		result, err = tx.Exec("insert into nest_user.account(is_valid) values(?)", i)
		if err != nil {
			return err
		}
		var uid int64
		uid, err = result.LastInsertId()

		// 插入用户库——档案表
		_, err = tx.Exec("insert into nest_user.profile(id, kind, nickname) values(?, 1, ?)", uid, "测试小号"+strconv.Itoa(i))
		if err != nil {
			return err
		}

		// 模拟故障
		// 插入管理库——管理员与用户关系表
		_, err = tx.Exec("insert into nest_admin.admin_user(user_id, admin_id) values(?, 20)", uid)
		if err != nil {
			return err
		}

		fmt.Println("新增用户", uid)
	}

	return nil
}
