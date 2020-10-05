package sqlx

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

type mockedConn struct {
	query string
	args  []interface{}
}

func (c *mockedConn) Query(dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

func (c *mockedConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	c.query = query
	c.args = args
	fmt.Printf("Query: %s Args: %s\n", query, args)
	return nil, nil
}

func (c *mockedConn) Transact(fn TransactFn) error {
	panic("implement me")
}

func TestBulkInserter_Insert(t *testing.T) {
	runSqlTest(t, func(conn Conn) {
		//var conn mockedConn
		inserter, err := NewBulkInserter(conn, `INSERT INTO classroom_dau(classroom, user, count) VALUES(?, ?, ?)`)
		assert.Nil(t, err)

		for i := 0; i < 5; i++ {
			assert.Nil(t, inserter.Insert("class_"+strconv.Itoa(i), "user_"+strconv.Itoa(i), i))
		}
		inserter.Flush()
		//assert.Equal(t, `INSERT INTO classroom_dau(classroom, user, count) VALUES `+
		//	`('class_0', 'user_0', 0),('class_1', 'user_1', 1),`+
		//	`('class_2', 'user_2', 2),('class_3', 'user_3', 3),('class_4', 'user_4', 4)`,
		//	conn.query)
		//assert.Nil(t, conn.args)
	})
}

func TestBulkInserter_Suffix(t *testing.T) {
	//runSqlTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
	runSqlTest(t, func(conn Conn) {
		//var conn mockedConn
		inserter, err := NewBulkInserter(conn, `INSERT INTO classroom_dau(classroom, user, count) VALUES`+
			`(?, ?, ?) ON DUPLICATE KEY UPDATE is_overtime=VALUES(is_overtime)`)
		assert.Nil(t, err)

		for i := 0; i < 50; i++ {
			assert.Nil(t, inserter.Insert("class_"+strconv.Itoa(i), "user_"+strconv.Itoa(i), i))
		}
		inserter.Flush()
		//assert.Equal(t, `INSERT INTO classroom_dau(classroom, user, count) VALUES `+
		//	`('class_0', 'user_0', 0),('class_1', 'user_1', 1),`+
		//	`('class_2', 'user_2', 2),('class_3', 'user_3', 3),('class_4', 'user_4', 4) ON DUPLICATE KEY UPDATE is_overtime=VALUES(is_overtime)`,
		//	conn.query)
		//assert.Nil(t, conn.args)
	})
}

//func runSqlTest(t *testing.T, fn func(db *sql.DB, mock sqlmock.Sqlmock)) {
func runSqlTest(t *testing.T, fn func(db Conn)) {
	//logx.Disable()
	//db, mock, err := sqlmock.New()
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)

	//if err != nil {
	//	t.Fatalf("打开数据库连接错误: %s", err)
	//}
	//defer db.Close()

	//fn(db, mock)
	fn(db)

	//if err := mock.ExpectationsWereMet(); err != nil {
	//	t.Errorf("存在为满足异常: %s", err)
	//}
}
