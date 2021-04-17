package contracts

import (
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type PatchInquiryStatusByUUIDParams struct {
	UUID          string
	InquiryStatus models.InquiryStatus
}

type InquiryResult struct {
	models.ServiceInquiry
	Username  string         `json:"username"`
	UserUuid  string         `json:"user_uuid"`
	AvatarUrl sql.NullString `json:"avatar_url"`
}

type InquiryDAOer interface {
	GetInquiryByUuid(iqUuid string, fields ...string) (*InquiryResult, error)
	PatchInquiryStatusByUUID(params PatchInquiryStatusByUUIDParams) error
	GetInquirerByInquiryUUID(uuid string, fields ...string) (*models.User, error)
}
