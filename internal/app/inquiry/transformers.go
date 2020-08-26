package inquiry

import (
	"fmt"
	"strconv"
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/shopspring/decimal"
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
	Budget        string    `json:"budget"`
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

type TransformedService struct {
	Uuid          string              `json:"uuid"`
	ServiceStatus string              `json:"service_status"`
	ServiceType   string              `json:"service_type"`
	User          TransformedInquirer `json:"inquirer"`
}

type TransformedInquirer struct {
	Uuid        string `json:"uuid"`
	Username    string `json:"username"`
	PremiumType string `json:"premium_type"`
}

func (t *InquiryTransform) TransformService(m models.Service, iqer models.User) TransformedService {
	return TransformedService{
		Uuid:          m.Uuid.String(),
		ServiceStatus: string(m.ServiceStatus),
		ServiceType:   string(m.ServiceType),
		User: TransformedInquirer{
			Uuid:        iqer.Uuid,
			Username:    iqer.Username,
			PremiumType: string(iqer.PremiumType),
		},
	}
}

type TransformedPickupInquiry struct {
	TransformedInquiry
	Inquirer TransformedInquirer `json:"inquirer"`
}

func (t *InquiryTransform) TransformPickupInquiry(iq models.ServiceInquiry, iqer models.User) TransformedPickupInquiry {
	tiq := t.TransformInquiry(iq)

	return TransformedPickupInquiry{
		tiq,
		TransformedInquirer{
			Uuid:        iqer.Uuid,
			Username:    iqer.Username,
			PremiumType: string(iqer.PremiumType),
		},
	}

}

type TransformedGirlApproveInquiry struct {
	TransformedInquiry
	Price           string    `json:"price"`
	Duration        int32     `json:"duration"`
	AppointmentTime time.Time `json:"appointment_time"`
	Lat             string    `json:"lat"`
	Lng             string    `json:"lng"`
}

func (t *InquiryTransform) TransformGirlApproveInquiry(iq models.ServiceInquiry) (*TransformedGirlApproveInquiry, error) {
	tiq := t.TransformInquiry(iq)

	tPrice, err := strconv.ParseFloat(
		iq.Price.String,
		64,
	)

	if err != nil {
		return nil, err
	}

	latDec, err := decimal.NewFromString(iq.Lat.String)

	if err != nil {
		return nil, err
	}

	lngDec, err := decimal.NewFromString(iq.Lng.String)

	if err != nil {
		return nil, err
	}

	return &TransformedGirlApproveInquiry{
		tiq,
		fmt.Sprintf("%.2f", util.RoundDown2Deci(tPrice)),
		iq.Duration.Int32,
		iq.AppointmentTime.Time,
		latDec.String(),
		lngDec.String(),
	}, nil
}
