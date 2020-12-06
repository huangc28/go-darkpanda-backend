package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type ImageDAOer interface {
	WithTx(tx db.Conn) ImageDAOer
	GetImagesByUserID(ID int) ([]models.Image, error)
}
