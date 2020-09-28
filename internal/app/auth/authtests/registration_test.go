package authtests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRegistrationTestSuite struct {
	suite.Suite
	sendRequest util.SendRequest
}

func (suite *UserRegistrationTestSuite) SetupSuite() {
	manager.NewDefaultManager()
	suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
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

	params := &url.Values{}
	params.Add("refer_code", ReferCode)
	params.Add("username", "somename")
	params.Add("gender", "female")

	req, err := http.NewRequest("POST", "/v1/register", strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	assert.Equal(suite.T(), dbUser.PhoneVerified, false) // the value of phone_verified is false
	assert.Equal(suite.T(), resUser.Uuid, dbUser.Uuid)
}

func (suite *UserRegistrationTestSuite) TestSendVerifyCodeViaTwilioSuccess() {
	// create a new user that wants to proceed phone verification process
	q := models.New(db.GetDB())

	usr, err := q.CreateUser(context.Background(), models.CreateUserParams{
		Username:      "someguy",
		Uuid:          "someuuid",
		PhoneVerified: false,
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
	data := url.Values{}
	data.Set("uuid", usr.Uuid)
	data.Set("mobile", "+886988272727")

	//bodyByte, _ := json.Marshal(&body)
	req, err := http.NewRequest("POST", "/v1/send-verify-code", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

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
	assert.Equal(suite.T(), respStruct.Uuid, usr.Uuid)
}

func (suite *UserRegistrationTestSuite) TestVerifyPhoneSuccess() {
	ctx := context.Background()

	newUserParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatalf("failed to generate test user param %s", err.Error())
	}

	// ------------------- tweak on user params to create new user -------------------
	newUserParams.PhoneVerified = false
	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatalf("failed to create test user %s", err.Error())
	}

	// ------------------- request phone verify API -------------------
	params := url.Values{}
	params.Add("uuid", newUser.Uuid)
	params.Add("verify_code", newUser.PhoneVerifyCode.String)
	params.Add("mobile", "+886988272727")

	req, err := http.NewRequest("POST", "/v1/verify-phone", strings.NewReader(params.Encode()))

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	router := app.StartApp(gin.Default())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Result().StatusCode; status != http.StatusOK {
		suite.T().Fatalf("Failed to verify code %s", string(rr.Body.Bytes()))
	}

	// ------------------- retrieve from DB makesure phone is verified -------------------
	dbuser, _ := q.GetUserByVerifyCode(ctx, newUser.PhoneVerifyCode)

	assert.Equal(suite.T(), dbuser.PhoneVerified, true)
	// ------------------- response has jwt token -------------------
	dec := json.NewDecoder(rr.Body)
	rBody := struct {
		JwtToken string `json:"jwt"`
	}{}
	if err := dec.Decode(&rBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(suite.T(), rBody.JwtToken)
}

func (suite *UserRegistrationTestSuite) TestRevokeJwtSuccess() {
	// ------------------- generate jwt token -------------------
	jwt, err := jwtactor.CreateToken("someuuid", config.GetAppConf().JwtSecret)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- create request -------------------
	body := struct {
		Jwt string `json:"jwt"`
	}{jwt}

	headers := make(map[string]string)
	resp, err := suite.sendRequest("POST", "/v1/revoke-jwt", body, headers)

	if resp.Result().StatusCode != http.StatusOK {
		suite.T().Fatalf("Failed to revoke jwt token")
	}

	// ------------------- check in redis if jwt exists -------------------
	ctx := context.Background()
	rds := db.GetRedis()
	isMember, err := rds.SIsMember(ctx, auth.INVALIDATE_TOKEN_REDIS_KEY, jwt).Result()

	if err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), isMember, true)
}

func TestRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserRegistrationTestSuite))
}
