package contracts

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type InquiryDAOer interface {
	GetInquiryByUuid(iqUuid string, fields ...string) (*models.ServiceInquiry, error)
}
