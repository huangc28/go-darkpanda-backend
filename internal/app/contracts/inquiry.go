package contracts

import (
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/db"
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

type GetInquiriesParams struct {
	UserID      int
	Offset      int
	PerPage     int
	InquiryType models.InquiryType
	Statuses    []models.InquiryStatus
}

type InquiryDAOer interface {
	WithTx(tx db.Conn) InquiryDAOer
	CheckHasActiveRandomInquiryByID(id int64) (bool, error)
	GetInquiries(p GetInquiriesParams) ([]*models.InquiryInfo, error)
	GetInquiryByUuid(iqUuid string, fields ...string) (*InquiryResult, error)
	HasMoreInquiries(offset int, perPage int) (bool, error)
	AskingInquiry(pickerID, inquiryID int64) (*models.ServiceInquiry, error)
	PatchInquiryStatusByUUID(params PatchInquiryStatusByUUIDParams) error
	GetInquirerByInquiryUUID(uuid string, fields ...string) (*models.User, error)
	PatchInquiryByInquiryUUID(params models.PatchInquiryParams) (*models.ServiceInquiry, error)
	GetActiveInquiry(inquirerId int) (*models.ActiveInquiry, error)
	GetInquiryByChannelUuid(channelUuid string) (*models.ServiceInquiry, error)
}
