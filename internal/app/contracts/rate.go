package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
)

type RateDAOer interface {
	WithTx(tx db.Conn) RateDAOer
}
