// Code generated by sqlc. DO NOT EDIT.

package models

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.createInquiryStmt, err = db.PrepareContext(ctx, createInquiry); err != nil {
		return nil, fmt.Errorf("error preparing query CreateInquiry: %w", err)
	}
	if q.createRefcodeStmt, err = db.PrepareContext(ctx, createRefcode); err != nil {
		return nil, fmt.Errorf("error preparing query CreateRefcode: %w", err)
	}
	if q.createUserStmt, err = db.PrepareContext(ctx, createUser); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUser: %w", err)
	}
	if q.getInquiryByInquirerIDStmt, err = db.PrepareContext(ctx, getInquiryByInquirerID); err != nil {
		return nil, fmt.Errorf("error preparing query GetInquiryByInquirerID: %w", err)
	}
	if q.getReferCodeInfoByRefcodeStmt, err = db.PrepareContext(ctx, getReferCodeInfoByRefcode); err != nil {
		return nil, fmt.Errorf("error preparing query GetReferCodeInfoByRefcode: %w", err)
	}
	if q.getUserByUsernameStmt, err = db.PrepareContext(ctx, getUserByUsername); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByUsername: %w", err)
	}
	if q.getUserByUuidStmt, err = db.PrepareContext(ctx, getUserByUuid); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByUuid: %w", err)
	}
	if q.getUserByVerifyCodeStmt, err = db.PrepareContext(ctx, getUserByVerifyCode); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByVerifyCode: %w", err)
	}
	if q.updateVerifyCodeByIdStmt, err = db.PrepareContext(ctx, updateVerifyCodeById); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateVerifyCodeById: %w", err)
	}
	if q.updateVerifyStatusByIdStmt, err = db.PrepareContext(ctx, updateVerifyStatusById); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateVerifyStatusById: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.createInquiryStmt != nil {
		if cerr := q.createInquiryStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createInquiryStmt: %w", cerr)
		}
	}
	if q.createRefcodeStmt != nil {
		if cerr := q.createRefcodeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createRefcodeStmt: %w", cerr)
		}
	}
	if q.createUserStmt != nil {
		if cerr := q.createUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createUserStmt: %w", cerr)
		}
	}
	if q.getInquiryByInquirerIDStmt != nil {
		if cerr := q.getInquiryByInquirerIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getInquiryByInquirerIDStmt: %w", cerr)
		}
	}
	if q.getReferCodeInfoByRefcodeStmt != nil {
		if cerr := q.getReferCodeInfoByRefcodeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getReferCodeInfoByRefcodeStmt: %w", cerr)
		}
	}
	if q.getUserByUsernameStmt != nil {
		if cerr := q.getUserByUsernameStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserByUsernameStmt: %w", cerr)
		}
	}
	if q.getUserByUuidStmt != nil {
		if cerr := q.getUserByUuidStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserByUuidStmt: %w", cerr)
		}
	}
	if q.getUserByVerifyCodeStmt != nil {
		if cerr := q.getUserByVerifyCodeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserByVerifyCodeStmt: %w", cerr)
		}
	}
	if q.updateVerifyCodeByIdStmt != nil {
		if cerr := q.updateVerifyCodeByIdStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateVerifyCodeByIdStmt: %w", cerr)
		}
	}
	if q.updateVerifyStatusByIdStmt != nil {
		if cerr := q.updateVerifyStatusByIdStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateVerifyStatusByIdStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                            DBTX
	tx                            *sql.Tx
	createInquiryStmt             *sql.Stmt
	createRefcodeStmt             *sql.Stmt
	createUserStmt                *sql.Stmt
	getInquiryByInquirerIDStmt    *sql.Stmt
	getReferCodeInfoByRefcodeStmt *sql.Stmt
	getUserByUsernameStmt         *sql.Stmt
	getUserByUuidStmt             *sql.Stmt
	getUserByVerifyCodeStmt       *sql.Stmt
	updateVerifyCodeByIdStmt      *sql.Stmt
	updateVerifyStatusByIdStmt    *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                            tx,
		tx:                            tx,
		createInquiryStmt:             q.createInquiryStmt,
		createRefcodeStmt:             q.createRefcodeStmt,
		createUserStmt:                q.createUserStmt,
		getInquiryByInquirerIDStmt:    q.getInquiryByInquirerIDStmt,
		getReferCodeInfoByRefcodeStmt: q.getReferCodeInfoByRefcodeStmt,
		getUserByUsernameStmt:         q.getUserByUsernameStmt,
		getUserByUuidStmt:             q.getUserByUuidStmt,
		getUserByVerifyCodeStmt:       q.getUserByVerifyCodeStmt,
		updateVerifyCodeByIdStmt:      q.updateVerifyCodeByIdStmt,
		updateVerifyStatusByIdStmt:    q.updateVerifyStatusByIdStmt,
	}
}
