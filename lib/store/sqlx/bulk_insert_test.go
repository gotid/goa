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
	//logx.Disable()
	//logx.SetLevel(logx.ErrorLevel)
	//runSqlTest(t, func(conn *command.DB, mock sqlmock.Sqlmock) {
	runSqlTest(t, func(conn Conn) {
		//var conn mockedConn
		inserter, err := NewBulkInserter(conn, `INSERT INTO nest_content_online.content_feed(user_id, content) VALUES`+
			`(?, ?) ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`)
		assert.Nil(t, err)

		//inserter.SetRequestHandler(func(result command.Result, err error) {
		//	affected, err := result.RowsAffected()
		//	insertId, err := result.LastInsertId()
		//	logx.Infof("协程数量：%d, 影响行数：%d, 返回编号：%d", runtime.NumGoroutine(), affected, insertId)
		//})

		for i := 0; i < 10; i++ {
			//assert.Nil(t, inserter.Insert(rand.Intn(218-6)+6, "70多国在联合国发言支持中方立场"+strconv.Itoa(i)))
			assert.Nil(t, inserter.Insert(1, "为中国正名！联合国公布，中国排名第一，让美国出乎意料"+strconv.Itoa(i)))
		}
		inserter.Flush()
		//assert.Equal(t, `INSERT INTO classroom_dau(classroom, user, count) VALUES `+
		//	`('class_0', 'user_0', 0),('class_1', 'user_1', 1),`+
		//	`('class_2', 'user_2', 2),('class_3', 'user_3', 3),('class_4', 'user_4', 4) ON DUPLICATE KEY UPDATE is_overtime=VALUES(is_overtime)`,
		//	conn.query)
		//assert.Nil(t, conn.args)
	})
}

//func runSqlTest(t *testing.T, fn func(conn *sql.DB, mock sqlmock.Sqlmock)) {
func runSqlTest(t *testing.T, fn func(db Conn)) {
	//logx.Disable()
	//conn, mock, err := sqlmock.New()
	dataSourceName := "root:asdfasdf@tcp(192.168.0.166:3306)/nest_label?parseTime=true"
	db := NewMySQL(dataSourceName)

	//if err != nil {
	//	t.Fatalf("打开数据库连接错误: %s", err)
	//}
	//defer conn.Close()

	//fn(conn, mock)
	fn(db)

	//if err := mock.ExpectationsWereMet(); err != nil {
	//	t.Errorf("存在为满足异常: %s", err)
	//}
}
