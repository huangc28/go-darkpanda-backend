package db

import (
	"net/http"

	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/jmoiron/sqlx"
)

type TxFunc func(tx *sqlx.Tx) (error, interface{})

func Transact(db *sqlx.DB, txFunc TxFunc) (error, interface{}) {
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

type FormatResp struct {
	Err            error
	ErrCode        string
	HttpStatusCode int
	Response       interface{}
}

type TxFuncFormatResp func(tx *sqlx.Tx) FormatResp

func TransactWithFormatStruct(db *sqlx.DB, txFunc TxFuncFormatResp) FormatResp {
	tx, err := db.Beginx()

	if err != nil {
		return FormatResp{
			Err:     err,
			ErrCode: apperr.FailedToBeginTx,
		}
	}

	fnResp := txFunc(tx)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if fnResp.Err != nil {
			tx.Rollback()

			// If http status code is not set, default to be `500`
			if fnResp.HttpStatusCode == 0 {
				fnResp.HttpStatusCode = http.StatusInternalServerError
			}
		} else {
			fnResp.Err = tx.Commit()

			if fnResp.Err != nil {
				fnResp.ErrCode = apperr.FailedToCommitTx
				fnResp.HttpStatusCode = http.StatusInternalServerError
			}
		}
	}()

	return fnResp
}
