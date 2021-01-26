package registertests

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/register"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRegistrationTestSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func (suite *UserRegistrationTestSuite) SetupSuite() {
	suite.assert = assert.New(suite.T())

	manager.
		NewDefaultManager(context.Background()).Run(func() {
		deps.Get().Run()
	})
}

func (suite *UserRegistrationTestSuite) BeforeTest(suiteName, testName string) {
	if testName == "TestSendVerifyCodeViaTwilioSuccess" {
		viper.Set("twilio.from", "+15005550006")
	}
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
	register.HandleRegister(c)
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
	register.HandleRegister(c)

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
