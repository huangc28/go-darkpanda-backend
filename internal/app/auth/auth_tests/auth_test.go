package authtests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	genverifycode "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/generate_verify_code"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserAuthTestSuite struct {
	suite.Suite
	sendRequest           util.SendRequest
	sendURLEncodedRequest util.SendUrlEncodedRequest
}

func (suite *UserAuthTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			suite.sendURLEncodedRequest = util.SendUrlEncodedRequestToApp(app.StartApp(gin.Default()))
			suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
		})
}

func (suite *UserAuthTestSuite) TestSendLoginVerifyCodeSuccess() {
	// create a registered user
	newUserParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatal(err)
	}

	newUserParams.PhoneVerified = true
	newUserParams.Mobile = sql.NullString{
		Valid:  true,
		String: "+886988272727",
	}

	q := models.New(db.GetDB())

	ctx := context.Background()
	user, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- create request -------------------
	params := url.Values{}
	params.Add("username", user.Username)

	rr, err := suite.sendURLEncodedRequest(
		"POST",
		"/v1/auth/send-verify-code",
		&params,
		make(map[string]string),
	)

	if err != nil {
		log.Fatal(err)
	}

	if rr.Result().StatusCode != http.StatusOK {
		suite.T().Fatal(string(rr.Body.Bytes()))
	}

	// ------------------- test cases -------------------
	assert := assert.New(suite.T())

	// assert key exists in redis
	redis := db.GetRedis()
	exists, err := redis.Exists(ctx, fmt.Sprintf(auth.LoginAuthenticatorHashKey, user.Uuid)).Result()

	if err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(1, int(exists))

	// assert response contains verify prefix and user uuid
	respStruct := &auth.TransformedSendLoginMobileVerifyCode{}
	if err := json.Unmarshal(rr.Body.Bytes(), respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(respStruct.UUID)
	assert.NotEmpty(respStruct.VerifyPrefix)
}

func (suite *UserAuthTestSuite) TestUserAttemptToLoginMultipleTimes() {
	suite.T().Skip("user attempt to login multiple times")
}

func (suite *UserAuthTestSuite) TestVerifyLoginCodeSuccess() {
	// Create a new user
	// Create a login record for that new user
	newUserParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatal(err)
	}

	newUserParams.PhoneVerified = true
	newUserParams.Mobile = sql.NullString{
		Valid:  true,
		String: "+886988272727",
	}

	q := models.New(db.GetDB())
	ctx := context.Background()
	user, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}
	verifyCode := genverifycode.GenVerifyCode()

	// AuthDAO
	dao := auth.NewAuthDao(db.GetRedis())
	dao.CreateLoginVerifyCode(ctx, verifyCode.BuildCode(), user.Uuid)

	// ------------------- create request -------------------
	params := &url.Values{}
	params.Add("mobile", "+886988272727")
	params.Add("verify_char", verifyCode.Chars)
	params.Add("verify_dig", strconv.Itoa(verifyCode.Dig))
	params.Add("uuid", user.Uuid)

	resp, err := suite.sendURLEncodedRequest(
		"POST",
		"/v1/verify-login-code",
		params,
		make(map[string]string),
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	respStruct := struct {
		JWT string `json:"jwt"`
	}{}

	if err := json.Unmarshal(resp.Body.Bytes(), &respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(suite.T(), respStruct.JWT)
}

func (suite *UserAuthTestSuite) TestRevokeJwtSuccess() {
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
	resp, err := suite.sendRequest("POST", "/v1/auth/revoke-jwt", body, headers)

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

func TestUserAuthTestSuite(t *testing.T) {
	suite.Run(t, new(UserAuthTestSuite))
}
