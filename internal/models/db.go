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
	if q.createRefcodeStmt, err = db.PrepareContext(ctx, createRefcode); err != nil {
		return nil, fmt.Errorf("error preparing query CreateRefcode: %w", err)
	}
	if q.createUserStmt, err = db.PrepareContext(ctx, createUser); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUser: %w", err)
	}
	if q.getReferCodeInfoByRefcodeStmt, err = db.PrepareContext(ctx, getReferCodeInfoByRefcode); err != nil {
		return nil, fmt.Errorf("error preparing query GetReferCodeInfoByRefcode: %w", err)
	}
	if q.getUserByUsernameStmt, err = db.PrepareContext(ctx, getUserByUsername); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByUsername: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
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
	createRefcodeStmt             *sql.Stmt
	createUserStmt                *sql.Stmt
	getReferCodeInfoByRefcodeStmt *sql.Stmt
	getUserByUsernameStmt         *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                            tx,
		tx:                            tx,
		createRefcodeStmt:             q.createRefcodeStmt,
		createUserStmt:                q.createUserStmt,
		getReferCodeInfoByRefcodeStmt: q.getReferCodeInfoByRefcodeStmt,
		getUserByUsernameStmt:         q.getUserByUsernameStmt,
	}
}
