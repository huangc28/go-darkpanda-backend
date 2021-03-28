package contracts

import (
	"context"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type UpdateUserInfoParams struct {
	AvatarURL       *string
	Nationality     *string
	Region          *string
	Age             *int
	Height          *float64
	Weight          *float64
	Description     *string
	BreastSize      *string
	PhoneVerifyCode *string
	PhoneVerified   *bool
	Uuid            string
	Mobile          string
}

type UserDAOer interface {
	GetUserInfoWithInquiryByUuid(ctx context.Context, uuid string, inquiryStatus models.InquiryStatus) (*models.UserWithInquiries, error)
	GetUserByUsername(username string, fields ...string) (*models.User, error)
	UpdateUserInfoByUuid(p UpdateUserInfoParams) (*models.User, error)
	GetUserByUuid(uuid string, fields ...string) (*models.User, error)
	GetUserByID(ID int64, fields ...string) (*models.User, error)
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFemaleByUuid(uuid string) (bool, error)
	GetUserImagesByUuid(uuid string, offset int, perPage int) ([]models.Image, error)
	WithTx(tx *sqlx.Tx) UserDAOer
}
