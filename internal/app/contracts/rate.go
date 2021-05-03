package contracts

import (
	"database/sql"
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
)

type GetUserRatingParams struct {
	ID        int            `json:"id"`
	Username  string         `json:"username"`
	AvatarUrl sql.NullString `json:"avatar_url"`
	Rating    int            `json:"rating"`
	Comments  sql.NullString `json:"comments"`
	CreatedAt time.Time      `json:"created_at"`
}

type RateDAOer interface {
	WithTx(tx db.Conn) RateDAOer
}
