package contracts

import (
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/db"
)

type GetUserBlockListParams struct {
	ID        int            `form:"id" json:"id"`
	UserId    int            `form:"user_id" json:"user_id"`
	Username  string         `form:"username" json:"username"`
	AvatarUrl sql.NullString `form:"avatar_url" json:"avatar_url"`
}

type BlockDAOer interface {
	WithTx(tx db.Conn) BlockDAOer
}
