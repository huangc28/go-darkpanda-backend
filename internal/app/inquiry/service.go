package inquiry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/shopspring/decimal"
	"github.com/teris-io/shortid"
)

type InquiryServicer interface {
}
type InquiryService struct {
	iqDao contracts.InquiryDAOer
	q     *models.Queries
	df    *darkfirestore.DarkFirestore
}

func NewService(iqDao contracts.InquiryDAOer, q *models.Queries, df *darkfirestore.DarkFirestore) *InquiryService {
	return &InquiryService{
		iqDao: iqDao,
		q:     q,
		df:    df,
	}
}

// CreateDirectInquiry male user wants to chat with a specific female.
type CreateDirectInquiryParams struct {
	InquirerUUID string
	InquirerID   int32

	PickerID        int32
	Budget          float64
	ServiceType     string
	AppointmentTime time.Time
	ServiceDuration int
	Address         string
}

func (s *InquiryService) CreateDirectInquiry(ctx context.Context, p CreateDirectInquiryParams) (*models.ServiceInquiry, error) {
	sid, _ := shortid.Generate()

	iq, err := s.q.CreateInquiry(ctx, models.CreateInquiryParams{
		Uuid: sid,
		InquirerID: sql.NullInt32{
			Valid: true,
			Int32: p.InquirerID,
		},
		PickerID: sql.NullInt32{
			Valid: true,
			Int32: p.PickerID,
		},

		Budget:        decimal.NewFromFloat(p.Budget).String(),
		ServiceType:   models.ServiceType(p.ServiceType),
		InquiryStatus: models.InquiryStatusInquiring,
		ExpiredAt: sql.NullTime{
			Time:  time.Now().Add(InquiryDuration),
			Valid: true,
		},
		AppointmentTime: sql.NullTime{
			Valid: true,

			// Convert appointment time to UTC to be consistent.
			Time: p.AppointmentTime.UTC(),
		},
		Duration: sql.NullInt32{
			Valid: true,
			Int32: int32(p.ServiceDuration),
		},
		Address: sql.NullString{
			Valid:  true,
			String: p.Address,
		},
		InquiryType: models.InquiryTypeDirect,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create direct inquiry %s: ", err.Error())
	}

	// Create new inquiry on firebase.
	_, _, err = s.df.CreateInquiringUser(
		ctx, darkfirestore.CreateInquiringUserParams{
			InquiryUuid:  iq.Uuid,
			InquirerUuid: p.InquirerUUID,
			InquiryType:  string(models.InquiryTypeDirect),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create direct inquiry on firestore %s: ", err.Error())
	}

	return &iq, nil
}

type CreateRandomInquiryParams struct {
	InquirerUUID string
	InquirerID   int32

	Budget          float64
	ServiceType     string
	AppointmentTime time.Time
	ServiceDuration int
	Address         string
}

func (s *InquiryService) CreateRandomInquiry(ctx context.Context, p CreateRandomInquiryParams) (*models.ServiceInquiry, error) {
	sid, _ := shortid.Generate()

	iq, err := s.q.CreateInquiry(
		ctx,
		models.CreateInquiryParams{
			Uuid: sid,
			InquirerID: sql.NullInt32{
				Int32: int32(p.InquirerID),
				Valid: true,
			},
			Budget:        decimal.NewFromFloat(p.Budget).String(),
			ServiceType:   models.ServiceType(p.ServiceType),
			InquiryStatus: models.InquiryStatusInquiring,
			ExpiredAt: sql.NullTime{
				Time:  time.Now().Add(InquiryDuration),
				Valid: true,
			},
			AppointmentTime: sql.NullTime{
				Valid: true,

				// Convert appointment time to UTC to be consistent.
				Time: p.AppointmentTime.UTC(),
			},
			Duration: sql.NullInt32{
				Valid: true,
				Int32: int32(p.ServiceDuration),
			},
			Address: sql.NullString{
				Valid:  true,
				String: p.Address,
			},
			InquiryType: models.InquiryTypeRandom,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create random inquiry %s: ", err.Error())
	}

	_, _, err = s.df.CreateInquiringUser(
		ctx, darkfirestore.CreateInquiringUserParams{
			InquiryUuid:  iq.Uuid,
			InquirerUuid: p.InquirerUUID,
			InquiryType:  string(models.InquiryTypeRandom),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create inquiring user on firestore %s: ", err.Error())
	}

	return &iq, nil
}
