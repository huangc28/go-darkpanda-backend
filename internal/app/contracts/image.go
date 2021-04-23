package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type ImageDAOer interface {
	WithTx(tx *sqlx.Tx) ImageDAOer
	GetImagesByUserID(ID int) ([]models.Image, error)
	CreateImages(imagesParams []models.Image) error
}
