package contracts

import (
	"context"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type UpdateUserInfoParams struct {
	Uuid          string
	Mobile        *string
	AvatarURL     *string
	Nationality   *string
	Region        *string
	Age           *int
	Height        *float64
	Weight        *float64
	Description   *string
	BreastSize    *string
	PhoneVerified *bool
}

type GetGirlsParams struct {
	// InquirerID is used to retrieve any latest inquiry with the girl.
	InquirerID int
	Limit      int
	Offset     int
}

type GetServiceOptionParams struct {
	UserID int
}

type CreateServiceOptionParams struct {
	UserID          int
	ServiceOptionID int
}

type CreateServiceOptionsParams struct {
	Name               string
	Description        string
	Price              float64
	ServiceOptionsType string
}

type UserDAOer interface {
	GetUserInfoWithInquiryByUuid(ctx context.Context, uuid string, inquiryStatus models.InquiryStatus) (*models.UserWithInquiries, error)
	GetUserByUsername(username string, fields ...string) (*models.User, error)
	UpdateUserInfoByUuid(p UpdateUserInfoParams) (*models.User, error)
	GetUserByUuid(uuid string, fields ...string) (*models.User, error)
	GetUserByID(ID int64, fields ...string) (*models.User, error)
	GetRating(userID int) (*models.UserRating, error)
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFemaleByUuid(uuid string) (bool, error)
	GetUserImagesByUuid(uuid string, offset int, perPage int) ([]models.Image, error)
	DeleteUserImages(url string) error
	GetGirls(p GetGirlsParams) ([]*models.RandomGirl, error)
	WithTx(tx *sqlx.Tx) UserDAOer
	GetUserServiceOption(userID int) ([]models.UserServiceOptionData, error)
	CreateServiceOption(p CreateServiceOptionsParams) (*models.ServiceOption, error)
	CreateUserServiceOption(p CreateServiceOptionParams) (*models.UserServiceOption, error)
	DeleteUserServiceOption(userID int, serviceOptionID int) error
}
