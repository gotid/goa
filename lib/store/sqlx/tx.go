package sqlx

import (
	"database/sql"
	"fmt"
)

type (
	beginTxFn func(*sql.DB) (TxSession, error)

	TxSession interface {
		Session
		Commit() error
		Rollback() error
	}

	txSession struct {
		*sql.Tx
	}
)

func doTx(c *conn, beginTx beginTxFn, transact TransactFn) (err error) {
	db, err := getConn(c.driverName, c.dataSourceName)
	if err != nil {
		logConnError(c.dataSourceName, err)
		return err
	}

	var tx TxSession
	tx, err = beginTx(db)
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务恢复自 %v, 回滚失败: %v", p, e)
			} else {
				err = fmt.Errorf("事务恢复自 %v", p)
			}
		} else if err != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务失败: %s, 回滚失败: %s", err, e)
			}
		} else {
			err = tx.Commit()
		}
	}()

	return transact(tx)
}

func beginTx(db *sql.DB) (TxSession, error) {
	if tx, err := db.Begin(); err != nil {
		return nil, err
	} else {
		return txSession{Tx: tx}, nil
	}
}

// Query 带事务查询
func (tx txSession) Query(dest interface{}, query string, args ...interface{}) error {
	return doQuery(tx.Tx, func(rows *sql.Rows) error {
		return scan(rows, dest)
	}, query, args...)
}

// Exec 带事务执行
func (tx txSession) Exec(query string, args ...interface{}) (sql.Result, error) {
	return doExec(tx.Tx, query, args...)
}
