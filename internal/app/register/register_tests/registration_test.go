package registertests

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/referral"
	"github.com/huangc28/go-darkpanda-backend/internal/app/register"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRegistrationTestSuite struct {
	suite.Suite
	depCon container.Container
	assert *assert.Assertions
}

func (suite *UserRegistrationTestSuite) SetupSuite() {
	suite.assert = assert.New(suite.T())

	manager.
		NewDefaultManager(context.Background()).Run(func() {
		deps.Get().Run()
		suite.depCon = deps.Get().Container
	})
}

func (suite *UserRegistrationTestSuite) TestRegisterMissingParams() {
	const ReferCode = "somerefercode"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := &url.Values{}
	params.Set("refer_code", ReferCode)
	headers := make(map[string]string)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/register",
		params,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	register.RegisterHandler(c)
	apperr.HandleError()(c)

	resStruct := struct {
		ErrCode string `json:"err_code"`
	}{}

	json.Unmarshal(w.Body.Bytes(), &resStruct)

	suite.assert.Equal(
		apperr.FailedToValidateRegisterParams,
		resStruct,
	)
}

func (suite *UserRegistrationTestSuite) invitorInviteeProvider(ctx context.Context) []models.User {
	q := models.New(db.GetDB())
	mans := make([]models.User, 0)

	// creates 2 man, one invitor one invitee.
	for i := 0; i < 2; i++ {
		manParams, err := util.GenTestUserParams()

		if err != nil {
			suite.T().Fatal(err)
		}

		manParams.Gender = models.GenderMale
		man, err := q.CreateUser(ctx, *manParams)

		if err != nil {
			suite.T().Fatal(err)
		}

		mans = append(mans, man)
	}

	return mans
}

func (suite *UserRegistrationTestSuite) TestVerifyInvitorCodeFailedDueToExpired() {

	ctx := context.Background()

	// Create invitor and invitee.
	mans := suite.invitorInviteeProvider(ctx)

	invitor := mans[0]
	invitee := mans[1]

	// Create an expired referral code.
	q := models.New(db.GetDB())
	refCode, err := q.CreateRefcode(ctx, models.CreateRefcodeParams{
		InvitorID: int32(invitor.ID),
		InviteeID: sql.NullInt32{
			Valid: false,
		},
		RefCode:     "somerefcode",
		RefCodeType: models.RefCodeTypeInvitor,
		ExpiredAt: sql.NullTime{
			Valid: true,
			Time:  time.Now().AddDate(0, 0, -4),
		},
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	// Request the API.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := &url.Values{}
	params.Add("invitee_uuid", invitee.Uuid)
	params.Add("referral_code", refCode.RefCode)
	headers := make(map[string]string)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/register",
		params,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	referral.HandleVerifyReferralCode(c, suite.depCon)
	apperr.HandleError()(c)

	respStruct := struct {
		ErrCode string `json:"err_code"`
		ErrMsg  string `json:"err_msg"`
	}{}

	json.Unmarshal(w.Body.Bytes(), &respStruct)

	suite.assert.Equal(apperr.ReferralCodeExpired, respStruct.ErrCode)
}

func (suite *UserRegistrationTestSuite) TestRegisterApiSuccess() {
	q := models.New(db.GetDB())

	// Create invitor
	usr, err := q.CreateUser(context.Background(), models.CreateUserParams{
		Username:      "Bryan Huang",
		PhoneVerified: true,
		AuthSmsCode: sql.NullInt32{
			Int32: 3333,
			Valid: true,
		},
		Gender:      models.GenderFemale,
		PremiumType: models.PremiumTypeNormal,
		PremiumExpiryDate: sql.NullTime{
			Valid: false,
		},
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	// Create refer code data
	referCode := "somerefercode"
	_, err = q.CreateRefcode(context.Background(), models.CreateRefcodeParams{
		InvitorID: int32(usr.ID),
		InviteeID: sql.NullInt32{
			Valid: false,
		},
		RefCode:     referCode,
		RefCodeType: models.RefCodeTypeInvitor,
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	params := &url.Values{}
	params.Add("refer_code", referCode)
	params.Add("username", "somename")
	params.Add("gender", "female")
	headers := make(map[string]string)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/register",
		params,
		headers,
	)

	c.Request = req
	register.RegisterHandler(c)

	// Assertions
	resUser := register.TransformedUser{}
	json.Unmarshal(w.Body.Bytes(), &resUser)

	query := models.New(db.GetDB())
	dbUser, _ := query.GetUserByUsername(context.Background(), "somename")

	suite.assert.Equal(dbUser.Gender, models.GenderFemale)
	suite.assert.Equal(dbUser.PremiumType, models.PremiumTypeNormal)
	suite.assert.Equal(dbUser.PhoneVerified, false) // the value of phone_verified is false
	suite.assert.Equal(resUser.Uuid, dbUser.Uuid)
}

func TestUserRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserRegistrationTestSuite))
}
