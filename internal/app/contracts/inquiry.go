package contracts

import (
	"database/sql"
	"time"

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

type InquiryInfo struct {
	models.ServiceInquiry
	Inquirer models.User
}

type PatchInquiryParams struct {
	Uuid            string     `json:"uuid"`
	AppointmentTime *time.Time `json:"appointment_time"`
	Price           *float32   `json:"price"`
	Duration        *int       `json:"duration"`
	ServiceType     *string    `json:"service_type"`
	Address         *string    `json:"address"`
}

type InquiryDAOer interface {
	WithTx(tx db.Conn)
	CheckHasActiveInquiryByID(id int64) (bool, error)
	GetInquiries(offset int, perpage int, statuses ...models.InquiryStatus) ([]*InquiryInfo, error)
	GetInquiryByUuid(iqUuid string, fields ...string) (*InquiryResult, error)
	HasMoreInquiries(offset int, perPage int) (bool, error)
	AskingInquiry(pickerID, inquiryID int64) (*models.ServiceInquiry, error)
	PatchInquiryStatusByUUID(params PatchInquiryStatusByUUIDParams) error
	GetInquirerByInquiryUUID(uuid string, fields ...string) (*models.User, error)
	PatchInquiryByInquiryUUID(params PatchInquiryParams) (*models.ServiceInquiry, error)
	GetActiveInquiry(inquirerId int) (*models.ServiceInquiry, error)
}
