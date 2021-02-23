package inquirytests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AgreeToChatTestSuite struct {
	suite.Suite
	depCon container.Container
}

func (s *AgreeToChatTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			s.depCon = deps.Get().Container
		})
}

func (s *AgreeToChatTestSuite) TestAgreeToChatSuccess() {
	// Create an inquirer.
	inquirerParams, err := util.GenTestUserParams()

	if err != nil {
		s.T().Fatal(err)
	}

	inquirerParams.Gender = models.GenderMale

	q := models.New(db.GetDB())
	ctx := context.Background()

	inquirer, err := q.CreateUser(ctx, *inquirerParams)

	if err != nil {
		s.T().Fatal(err)
	}

	// Create an inquiry picker.
	pickerParams, err := util.GenTestUserParams()
	if err != nil {
		s.T().Fatal(err)
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
		s.T().Fatal(err)
	}

	// Create an inquiry with status `asking`.
	iqParams, err := util.GenTestInquiryParams(inquirer.ID)

	if err != nil {
		s.T().Fatal(err)
	}

	iqParams.InquiryStatus = models.InquiryStatusAsking
	iqParams.PickerID = sql.NullInt32{
		Valid: true,
		Int32: int32(picker.ID),
	}

	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		s.T().Fatal(err)
	}

	// Create an inquiry in firestore with status asking
	df := darkfirestore.Get()
	df.CreateInquiringUser(
		ctx,
		darkfirestore.CreateInquiringUserParams{
			InquiryUUID: iq.Uuid,
		},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeTestRequest(
		"POST",
		fmt.Sprintf("/v1/%s/agree-to-chat", iq.Uuid),
		&url.Values{},
		util.CreateJwtHeaderMap(
			inquirer.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		s.T().Fatal(err)
	}

	c.Request = req
	c.Params = append(
		c.Params,
		gin.Param{
			Key:   "inquiry_uuid",
			Value: iq.Uuid,
		},
	)

	jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	})(c)

	inquiry.AgreeToChatInquiryHandler(c, s.depCon)
	apperr.HandleError()(c)

	// ------------------- Assertions -------------------
	respStruct := inquiry.TransformedAgreePickupInquiry{}
	json.Unmarshal(w.Body.Bytes(), &respStruct)

	assert := assert.New(s.T())

	assert.Equal(picker.Username, respStruct.ServiceProvider.Username)
	assert.Equal(picker.AvatarUrl.String, respStruct.ServiceProvider.AvatarUrl)
	assert.Equal(picker.Uuid, respStruct.ServiceProvider.Uuid)
	assert.Equal(picker.Description.String, respStruct.ServiceProvider.Description)
	assert.NotEmpty(respStruct.ChannelUuid)
}

func TestAgreeToChatTestSuite(t *testing.T) {
	suite.Run(t, new(AgreeToChatTestSuite))
}
