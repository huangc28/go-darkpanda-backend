package authtests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRegistrationTestSuite struct {
	suite.Suite
}

func (suite *UserRegistrationTestSuite) SetupSuite() {
	manager.NewDefaultManager()
}

func (suite *UserRegistrationTestSuite) BeforeTest(suiteName, testName string) {
	if testName == "TestSendVerifyCodeViaTwilioSuccess" {
		viper.Set("twilio.from", "+15005550006")
	}
}

func (suite *UserRegistrationTestSuite) TestRegisterMissingParams() {
	const ReferCode = "somerefercode"
	body := struct {
		ReferCode string `json:"refer_code"`
	}{
		ReferCode,
	}

	bodyB, _ := json.Marshal(&body)
	req, err := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(bodyB))

	if err != nil {
		suite.T().Fatalf("[register_missing_params] failed to request registerAPI %s", err.Error())
	}

	router := app.StartApp(gin.Default())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(suite.T(), rr.Result().StatusCode, http.StatusBadRequest)
	r := json.NewDecoder(rr.Body)
	resStruct := struct {
		ErrCode string `json:"err_code"`
	}{}

	r.Decode(&resStruct)
	assert.Equal(suite.T(), resStruct.ErrCode, apperr.FailedToValidateRegisterParams)
}

func (suite *UserRegistrationTestSuite) TestRegisterApiSuccess() {
	q := models.New(db.GetDB())

	// Create invitor
	usr, err := q.CreateUser(context.Background(), models.CreateUserParams{
		Username: "Bryan Huang",
		PhoneVerified: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
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
	const ReferCode = "somerefercode"
	_, err = q.CreateRefcode(context.Background(), models.CreateRefcodeParams{
		InvitorID: int32(usr.ID),
		InviteeID: sql.NullInt32{
			Valid: false,
		},
		RefCode:     ReferCode,
		RefCodeType: models.RefCodeTypeInvitor,
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	body := struct {
		ReferCode string `json:"refer_code"`
		Username  string `json:"username"`
		Gender    string `json:"gender"`
	}{
		ReferCode,
		"somename",
		"female",
	}

	bodyB, _ := json.Marshal(&body)

	req, err := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(bodyB))

	if err != nil {
		suite.T().Fatal(err)
	}

	router := app.StartApp(gin.Default())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	dec := json.NewDecoder(rr.Body)
	var resUser auth.TransformedUser
	if err := dec.Decode(&resUser); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), resUser.Username, "somename")
	assert.Equal(suite.T(), resUser.Gender, "female")

	// retrieve users
	query := models.New(db.GetDB())
	dbUser, _ := query.GetUserByUsername(context.Background(), "somename")

	assert.Equal(suite.T(), dbUser.Gender, models.GenderFemale)
	assert.Equal(suite.T(), dbUser.PremiumType, models.PremiumTypeNormal)
	assert.Equal(suite.T(), dbUser.PhoneVerified.Bool, false) // the value of phone_verified is false
	assert.Equal(suite.T(), dbUser.PhoneVerified.Valid, true) // the value of phone_verified exists in table
	assert.Equal(suite.T(), resUser.Uuid, dbUser.Uuid)
}

func (suite *UserRegistrationTestSuite) TestSendVerifyCodeViaTwilioSuccess() {
	// create a new user that wants to proceed phone verification process
	q := models.New(db.GetDB())

	usr, err := q.CreateUser(context.Background(), models.CreateUserParams{
		Username: "someguy",
		Uuid:     "someuuid",
		PhoneVerified: sql.NullBool{
			Bool:  false,
			Valid: true,
		},
		AuthSmsCode: sql.NullInt32{
			Valid: false,
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

	body := struct {
		Uuid   string `json:"uuid"`
		Mobile string `json:"mobile"`
	}{
		usr.Uuid,
		"+886988272727",
	}

	bodyByte, _ := json.Marshal(&body)
	req, err := http.NewRequest("POST", "/v1/send-verify-code", bytes.NewBuffer(bodyByte))

	if err != nil {
		suite.
			T().
			Fatalf("[send_verify_code] failed to request registerAPI %s", err.Error())
	}

	router := app.StartApp(gin.Default())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	respStruct := struct {
		Uuid         string `json:"uuid"`
		VerifyPrefix string `json:"verify_prefix"`
		VerifySuffix int    `json:"verify_suffix"`
	}{}

	dec := json.NewDecoder(rr.Result().Body)
	dec.Decode(&respStruct)

	assert.NotEmpty(suite.T(), respStruct.VerifyPrefix)
	assert.NotEmpty(suite.T(), respStruct.VerifySuffix)
	assert.Equal(suite.T(), respStruct.Uuid, body.Uuid)
}

func TestRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserRegistrationTestSuite))
}
