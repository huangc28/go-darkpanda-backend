package test_helpers

import (
	"context"
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"

	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
)

type TestHelpers struct{}

func NewTestHelpers() *TestHelpers {
	return &TestHelpers{}
}

type CreateInquiryStatusParam struct {
	Status   models.InquiryStatus
	Picker   *models.User
	Inquirer *models.User
}

type CreateInquiryStatusUserResp struct {
	Inquiry  models.ServiceInquiry
	Picker   *models.User
	Inquirer *models.User
}

// CreateInquiryStatusUser is a helper fucntion that creates the
// following resources for testing purpose.
//   - New inquiry
//   - New inquirer (male user)
//   - New picker (female user)
//   - Set inquiry status specified by the user
func (*TestHelpers) CreateInquiryStatusUser(params CreateInquiryStatusParam) (*CreateInquiryStatusUserResp, error) {
	if params.Status == models.InquiryStatus("") {
		params.Status = models.InquiryStatusInquiring
	}

	var (
		picker   *models.User = params.Picker
		inquirer *models.User = params.Inquirer
	)

	q := models.New(db.GetDB())
	ctx := context.Background()

	// Not passing any inquirer, we need to create manually create an inquirer
	if inquirer == (*models.User)(nil) {
		inquirer = &models.User{}

		// Create an inquirer.
		inquirerParams, err := util.GenTestUserParams()

		if err != nil {
			return nil, err
		}

		inquirerParams.Gender = models.GenderMale

		*inquirer, err = q.CreateUser(ctx, *inquirerParams)

		if err != nil {
			return nil, err
		}
	}

	if picker == (*models.User)(nil) {
		picker = &models.User{}

		// Create an inquiry picker.
		pickerParams, err := util.GenTestUserParams()
		if err != nil {
			return nil, err
		}

		pickerParams.Gender = models.GenderFemale
		pickerParams.AvatarUrl = sql.NullString{
			Valid:  true,
			String: "http://darkpanda.com/somehornygirl/avatar.png",
		}

		*picker, err = q.CreateUser(ctx, *pickerParams)
		if err != nil {
			return nil, err
		}
	}

	// Create an inquiry with status `asking`.
	iqParams, err := util.GenTestInquiryParams(inquirer.ID)

	if err != nil {
		return nil, err
	}

	iqParams.InquiryStatus = params.Status
	iqParams.PickerID = sql.NullInt32{
		Valid: true,
		Int32: int32(picker.ID),
	}

	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		return nil, err
	}

	// Create an inquiry in firestore with status asking
	df := darkfirestore.Get()
	df.CreateInquiringUser(
		ctx,
		darkfirestore.CreateInquiringUserParams{
			InquiryUUID: iq.Uuid,
		},
	)

	return &CreateInquiryStatusUserResp{
		Inquiry:  iq,
		Picker:   picker,
		Inquirer: inquirer,
	}, nil
}

type CreateTestServiceParams struct {
	ServiceStatus models.ServiceStatus
	Picker        *models.User
	Inquirer      *models.User
}

type CreateTestServiceResponse struct {
	Inquiry  *models.ServiceInquiry
	Service  *models.Service
	Picker   *models.User
	Inquirer *models.User
}

func (th *TestHelpers) CreateTestService(p CreateTestServiceParams) (*CreateTestServiceResponse, error) {
	iqResp, err := th.CreateInquiryStatusUser(
		CreateInquiryStatusParam{
			Status:   models.InquiryStatusBooked,
			Picker:   p.Picker,
			Inquirer: p.Inquirer,
		},
	)

	if err != nil {
		return nil, err
	}

	srvParams, err := util.GenTestServiceParams(
		iqResp.Inquirer.ID,
		iqResp.Picker.ID,
		iqResp.Inquiry.ID,
	)

	srvParams.ServiceStatus = p.ServiceStatus

	if err != nil {
		return nil, err
	}

	q := models.New(db.GetDB())
	ctx := context.Background()
	srvModel, err := q.CreateService(ctx, *srvParams)

	if err != nil {
		return nil, err
	}

	return &CreateTestServiceResponse{
		Inquiry:  &iqResp.Inquiry,
		Service:  &srvModel,
		Picker:   iqResp.Picker,
		Inquirer: iqResp.Inquirer,
	}, nil

}

func (*TestHelpers) RemoveInquiry(ctx context.Context, inquiryUuid string) error {
	dfClient := darkfirestore.Get().Client
	_, err := dfClient.
		Collection("inquiries").
		Doc(inquiryUuid).
		Delete(ctx)

	return err
}
