package inquiry

import (
	"fmt"
	"strconv"
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	convertnullsql "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/convert_null_sql"
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

// TransformBookService response with the information of booked service and the information about
// the service provider.
// @TODO information of service provider should include provider image.
type TransformedServiceProvider struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
}

type TransformedBookedService struct {
	Uuid            string                     `json:"uuid"`
	ServiceProvider TransformedServiceProvider `json:"service_provider"`
	Price           string                     `json:"price"`
	Duration        int32                      `json:"duration"`
	AppointmentTime time.Time                  `json:"appointment_time"`
	Lng             string                     `json:"lng"`
	Lat             string                     `json:"lat"`
	ServiceType     string                     `json:"service_type"`
	CreatedAt       time.Time                  `json:"created_at"`
}

func (t *InquiryTransform) TransformBookedService(srv models.Service, userProvider models.User) *TransformedBookedService {
	tsrv := &TransformedBookedService{
		Uuid:            srv.Uuid.String(),
		Price:           srv.Price.String,
		Duration:        srv.Duration.Int32,
		AppointmentTime: srv.AppointmentTime.Time,
		Lng:             srv.Lng.String,
		Lat:             srv.Lat.String,
		ServiceType:     string(srv.ServiceType),
		CreatedAt:       srv.CreatedAt,
		ServiceProvider: TransformedServiceProvider{
			Uuid:     userProvider.Uuid,
			Username: userProvider.Username,
		},
	}

	return tsrv
}

// Transformed object for GET /v1/inquiries
type TransformedGetInquiryInquirer struct {
	Uuid        string `json:"uuid"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	Nationality string `json:"nationality"`
}

type TransformedGetInquiryWithInquirer struct {
	Uuid        string                        `json:"uuid"`
	Budget      float64                       `json:"budget"`
	ServiceType string                        `json:"service_type"`
	Price       *float64                      `json:"price"`
	Duration    int32                         `json:"duration"`
	Appointment time.Time                     `json:"appoinment_time"`
	Lng         *float32                      `json:"lng"`
	Lat         *float32                      `json:"lat"`
	Inquirer    TransformedGetInquiryInquirer `json:"inquirer"`
}

type TransformedInquiries struct {
	Inquiries []TransformedGetInquiryWithInquirer `json:"inquiries"`
	HasMore   bool                                `json:"has_more"`
}

func (t *InquiryTransform) TransformInquiryList(inquiryList []*InquiryInfo, hasMore bool) (TransformedInquiries, error) {
	trfedIqs := make([]TransformedGetInquiryWithInquirer, 0)
	for _, oi := range inquiryList {
		price, err := convertnullsql.ConvertSqlNullStringToFloat64(oi.Price)

		if err != nil {
			return TransformedInquiries{}, err
		}

		budget, err := strconv.ParseFloat(oi.Budget, 64)

		if err != nil {
			return TransformedInquiries{}, err
		}

		lng, err := convertnullsql.ConvertSqlNullStringToFloat32(oi.Lng)

		if err != nil {
			return TransformedInquiries{}, err
		}

		lat, err := convertnullsql.ConvertSqlNullStringToFloat32(oi.Lat)

		if err != nil {
			return TransformedInquiries{}, err
		}

		trfedIq := TransformedGetInquiryWithInquirer{
			Uuid:        oi.Uuid,
			Budget:      budget,
			ServiceType: oi.ServiceType.ToString(),
			Price:       price,
			Duration:    oi.Duration.Int32,
			Appointment: oi.AppointmentTime.Time,
			Lng:         lng,
			Lat:         lat,
			Inquirer: TransformedGetInquiryInquirer{
				Uuid:        oi.Inquirer.Uuid,
				Username:    oi.Inquirer.Username,
				AvatarURL:   oi.Inquirer.AvatarUrl.String,
				Nationality: oi.Inquirer.Nationality.String,
			},
		}

		trfedIqs = append(trfedIqs, trfedIq)
	}

	return TransformedInquiries{
		Inquiries: trfedIqs,
		HasMore:   hasMore,
	}, nil
}

type TransformedGetInquiry struct {
	Uuid        string    `json:"uuid"`
	Budget      float64   `json:"budget"`
	ServiceType string    `json:"service_type"`
	Price       *float64  `json:"price"`
	Duration    int32     `json:"duration"`
	Appointment time.Time `json:"appoinment_time"`
	Lng         *float32  `json:"lng"`
	Lat         *float32  `json:"lat"`
}

func (t *InquiryTransform) TransformGetInquiry(iq models.ServiceInquiry) (*TransformedGetInquiry, error) {

	budget, err := strconv.ParseFloat(iq.Budget, 64)

	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(iq.Budget, 64)

	if err != nil {
		return nil, err
	}

	lng, err := convertnullsql.ConvertSqlNullStringToFloat32(iq.Lng)

	if err != nil {
		return nil, err
	}

	lat, err := convertnullsql.ConvertSqlNullStringToFloat32(iq.Lat)

	if err != nil {
		return nil, err
	}

	return &TransformedGetInquiry{
		iq.Uuid,
		budget,
		iq.ServiceType.ToString(),
		&price,
		iq.Duration.Int32,
		iq.AppointmentTime.Time,
		lng,
		lat,
	}, nil
}
