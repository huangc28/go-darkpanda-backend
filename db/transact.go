package db

import (
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/jmoiron/sqlx"
)

func Transact(db *sqlx.DB, txFunc func(*sqlx.Tx) (error, interface{})) (error, interface{}) {
	tx, err := db.Beginx()

	if err != nil {
		return err, apperr.FailedToBeginTx
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err, extra := txFunc(tx)

	return err, extra
}
