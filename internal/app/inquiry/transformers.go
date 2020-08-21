package inquiry

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type InquiryTransformer interface {
	TransformInquiry(m models.ServiceInquiry) InquiryTransform
}

type InquiryTransform struct{}

func NewTransform() *InquiryTransform {
	return &InquiryTransform{}
}

type TransformedInquiry struct {
	Uuid          string    `json:"uuid"`
	Budget        float64   `json:"budget"`
	ServiceType   string    `json:"service_type"`
	InquiryStatus string    `json:"inquiry_status"`
	CreatedAt     time.Time `json:"created_at"`
}

func (t *InquiryTransform) TransformInquiry(m models.ServiceInquiry) TransformedInquiry {
	tiq := TransformedInquiry{
		Uuid:          m.Uuid,
		Budget:        m.Budget,
		ServiceType:   string(m.ServiceType),
		InquiryStatus: string(m.InquiryStatus),
		CreatedAt:     m.CreatedAt,
	}

	return tiq
}
