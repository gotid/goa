package sqlx

import (
	"database/sql"
	"fmt"
)

type (
	beginTxFn func(*sql.DB) (trans, error)

	trans interface {
		Session
		Commit() error
		Rollback() error
	}

	txSession struct {
		*sql.Tx
	}
)

func doTransaction(c *conn, beginTx beginTxFn, fn func(session Session) error) error {
	db, err := getConn(c.driverName, c.dataSourceName)
	if err != nil {
		logConnError(c.dataSourceName, err)
		return err
	}
	tx, err := beginTx(db)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("恢复自 %v, 回滚失败: %v", p, e)
			} else {
				err = fmt.Errorf("恢复自 %v", p)
			}
		} else if err != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务失败: %s, 回滚失败: %s", err, e)
			}
		} else {
			err = tx.Commit()
		}
	}()

	return fn(tx)
}

func beginTx(db *sql.DB) (trans, error) {
	if tx, err := db.Begin(); err != nil {
		return nil, err
	} else {
		return txSession{Tx: tx}, nil
	}
}

func (t txSession) Query(dest interface{}, query string, args ...interface{}) error {
	return doQuery(t.Tx, func(rows *sql.Rows) error {
		return scan(rows, dest)
	}, query, args...)
}

func (t txSession) Exec(query string, args ...interface{}) (sql.Result, error) {
	return doExec(t.Tx, query, args...)
}
