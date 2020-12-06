package contracts

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type PatchInquiryStatusByUUIDParams struct {
	UUID          string
	InquiryStatus models.InquiryStatus
}

type InquiryDAOer interface {
	GetInquiryByUuid(iqUuid string, fields ...string) (*models.ServiceInquiry, error)
	PatchInquiryStatusByUUID(params PatchInquiryStatusByUUIDParams) error
}
