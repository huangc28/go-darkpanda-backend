package helpers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"

	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
)

type CreateInquiryStatusParam struct {
	Status models.InquiryStatus
}

type CreateInquiryStatusUserResp struct {
	Inquiry  models.ServiceInquiry
	Picker   models.User
	Inquirer models.User
}

// CreateInquiryStatusUser is a helper fucntion that creates the
// following resources for testing purpose.
//   - New inquiry
//   - New inquirer (male user)
//   - New picker (female user)
//   - Set inquiry status specified by the user
func CreateInquiryStatusUser(t *testing.T, params CreateInquiryStatusParam) CreateInquiryStatusUserResp {
	if params.Status == models.InquiryStatus("") {
		params.Status = models.InquiryStatusInquiring
	}

	// Create an inquirer.
	inquirerParams, err := util.GenTestUserParams()

	if err != nil {
		t.Fatal(err)
	}

	inquirerParams.Gender = models.GenderMale

	q := models.New(db.GetDB())
	ctx := context.Background()

	inquirer, err := q.CreateUser(ctx, *inquirerParams)

	if err != nil {
		t.Fatal(err)
	}

	// Create an inquiry picker.
	pickerParams, err := util.GenTestUserParams()
	if err != nil {
		t.Fatal(err)
	}

	pickerParams.Gender = models.GenderFemale
	pickerParams.Username = "somehornygirl"
	pickerParams.Description = sql.NullString{
		Valid:  true,
		String: "iamahornygirlpoundmysnooch",
	}
	pickerParams.AvatarUrl = sql.NullString{
		Valid:  true,
		String: "http://darkpanda.com/somehornygirl/avatar.png",
	}

	picker, err := q.CreateUser(ctx, *pickerParams)
	if err != nil {
		t.Fatal(err)
	}

	// Create an inquiry with status `asking`.
	iqParams, err := util.GenTestInquiryParams(inquirer.ID)

	if err != nil {
		t.Fatal(err)
	}

	iqParams.InquiryStatus = params.Status
	iqParams.PickerID = sql.NullInt32{
		Valid: true,
		Int32: int32(picker.ID),
	}

	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		t.Fatal(err)
	}

	// Create an inquiry in firestore with status asking
	df := darkfirestore.Get()
	df.CreateInquiringUser(
		ctx,
		darkfirestore.CreateInquiringUserParams{
			InquiryUUID: iq.Uuid,
		},
	)

	return CreateInquiryStatusUserResp{
		Inquiry:  iq,
		Picker:   picker,
		Inquirer: inquirer,
	}
}

func RemoveInquiry(ctx context.Context, inquiryUuid string) error {
	dfClient := darkfirestore.Get().Client
	_, err := dfClient.
		Collection("inquiries").
		Doc(inquiryUuid).
		Delete(ctx)

	return err
}
