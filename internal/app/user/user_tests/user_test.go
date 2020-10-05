package usertests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserAPITestsSuite struct {
	suite.Suite
	sendRequest           util.SendRequest
	sendUrlEncodedRequest util.SendUrlEncodedRequest
}

func (suite *UserAPITestsSuite) SetupSuite() {
	manager.NewDefaultManager()
	tApp := app.StartApp(gin.Default())
	suite.sendRequest = util.SendRequestToApp(tApp)
	suite.sendUrlEncodedRequest = util.SendUrlEncodedRequestToApp(tApp)
}

func (suite *UserAPITestsSuite) TestGetMaleUserInfo() {
	// create a male user that has no related active inquiry
	ctx := context.Background()
	newUserParams, err := util.GenTestUserParams()
	newUserParams.Gender = models.GenderMale

	if err != nil {
		suite.T().Fatal(err)
	}

	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- Create jwt to request the API -------------------
	jwt, err := jwtactor.CreateToken(newUser.Uuid, config.GetAppConf().JwtSecret)

	if err != nil {
		suite.T().Fatal(err)
	}

	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", jwt)

	resp, _ := suite.sendRequest("GET", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases  -------------------
	assert.Equal(suite.T(), resp.Result().StatusCode, http.StatusOK)
	dec := json.NewDecoder(resp.Result().Body)
	respStruct := &user.TransformUserWithInquiryData{}
	if err := dec.Decode(&respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), respStruct.Username, newUser.Username)
	assert.Equal(suite.T(), respStruct.Gender, newUser.Gender)
	assert.Equal(suite.T(), respStruct.Uuid, newUser.Uuid)

	assert.Equal(suite.T(), len(respStruct.Inquiries), 0)

}

func (suite *UserAPITestsSuite) TestGetMaleUserInfoWithActiveInquiry() {
	// create a male user along with active inquiries
	ctx := context.Background()
	newUserParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatal(err)
	}

	newUserParams.Gender = models.GenderMale

	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	newInquiryParams, err := util.GenTestInquiryParams(newUser.ID)
	if err != nil {
		suite.T().Fatal(err)
	}
	newInquiryParams.InquiryStatus = models.InquiryStatusInquiring
	if _, err := q.CreateInquiry(ctx, *newInquiryParams); err != nil {
		suite.T().Fatalf("Failed to create inquiry %s", err.Error())
	}

	// ------------------- Create jwt to request the API -------------------
	jwt, err := jwtactor.CreateToken(newUser.Uuid, config.GetAppConf().JwtSecret)

	if err != nil {
		suite.T().Fatal(err)
	}

	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", jwt)
	resp, _ := suite.sendRequest("GET", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases -------------------
	respStruct := &user.TransformUserWithInquiryData{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), len(respStruct.Inquiries), 1)
}

func (suite *UserAPITestsSuite) TestGetFemaleUserInfo() {
	ctx := context.Background()
	userParams, _ := util.GenTestUserParams()
	userParams.Gender = models.GenderFemale

	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *userParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// request the api

	header := util.CreateJwtHeaderMap(
		newUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, _ := suite.sendRequest("GET", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases -------------------
	respStruct := &user.TransformedUser{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(respStruct.Uuid, newUser.Uuid)
	assert.Equal(respStruct.AvatarUrl, newUser.AvatarUrl.String)
	assert.Equal(respStruct.Username, newUser.Username)
	assert.Equal(respStruct.Gender, newUser.Gender)
}

func (suite *UserAPITestsSuite) TestPutUserInfoSuccess() {
	// ------------------- create male user -------------------
	ctx := context.Background()
	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = models.GenderMale
	q := models.New(db.GetDB())
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- send API -------------------
	headers := util.CreateJwtHeaderMap(
		maleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	body := struct {
		AvatarURL   string  `json:"avatar_url"`
		Nationality string  `json:"nationality"`
		Region      string  `json:"region"`
		Height      float32 `json:"height"`
		Weight      float32 `json:"weight"`
		Age         int     `json:"age"`
		Description string  `json:"description"`
		BreastSize  string  `json:"breast_size"`
	}{
		AvatarURL:   "https://somecloud.com/a.png",
		Nationality: "Taiwan",
		Region:      "Taipei",
		Height:      170,
		Weight:      68,
		Age:         18,
		Description: "I am rich and have big dick",
		BreastSize:  "z",
	}

	res, err := suite.sendRequest(
		"PUT",
		fmt.Sprintf("/v1/users/%s", maleUser.Uuid),
		body,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	if res.Result().StatusCode != http.StatusOK {
		bbody, _ := ioutil.ReadAll(res.Result().Body)
		suite.T().Fatalf("%s", string(bbody))
	}

	// ------------------- test asserts -------------------
	assert := assert.New(suite.T())
	dec := json.NewDecoder(res.Result().Body)
	resBody := struct {
		AvatarURL   *string `json:"avatar_url"`
		Nationality *string `json:"nationality"`
		Age         *int    `json:"age"`
	}{}

	if err := dec.Decode(&resBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(*resBody.AvatarURL, body.AvatarURL)
	assert.Equal(*resBody.Nationality, "Taiwan")
	assert.Equal(*resBody.Age, 18)

	// ------------------- test nil update -------------------
	resTwo, err := suite.sendRequest(
		"PUT",
		fmt.Sprintf("/v1/users/%s", maleUser.Uuid),
		struct {
			Age int `json:"age"`
		}{
			19,
		},
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	resTwoBody := struct {
		AvatarURL   *string `json:"avatar_url"`
		Nationality *string `json:"nationality"`
		Age         *int    `json:"age"`
	}{}
	dec = json.NewDecoder(resTwo.Result().Body)
	if err := dec.Decode(&resTwoBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(*resBody.AvatarURL, body.AvatarURL)
	assert.Equal(*resBody.Nationality, "Taiwan")
	assert.Equal(*resTwoBody.Age, 19)
}

func (suite *UserAPITestsSuite) TestGetUserProfileByUuid() {
	// create a female user
	ctx := context.Background()
	q := models.New(db.GetDB())

	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = "female"
	femaleUserParams.PhoneVerified = true
	femaleUserParams.Mobile = sql.NullString{
		Valid:  true,
		String: "+886988272727",
	}

	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

	if err != nil {
		suite.T().Fatalf("create female user %s", err.Error())
	}

	// create a male user
	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = "male"
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	maleUser.PhoneVerified = true
	maleUser.Mobile = sql.NullString{
		Valid:  true,
		String: "+886986900050",
	}

	if err != nil {
		suite.T().Fatalf("create male user %s", err.Error())
	}

	// female user wants to view the profile of a male user
	headers := util.CreateJwtHeaderMap(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, err := suite.sendUrlEncodedRequest(
		"GET",
		fmt.Sprintf("/v1/users/%s", maleUser.Uuid),
		&url.Values{},
		headers,
	)

	if resp.Code != http.StatusOK {
		suite.T().Fatal(err)
	}

	// ------------------- test cases -------------------
	respStruct := &user.TransformedMaleUser{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(maleUser.Uuid, respStruct.Uuid)
	assert.Equal(maleUser.Username, respStruct.Username)
}

func TestUserAPISuite(t *testing.T) {
	suite.Run(t, new(UserAPITestsSuite))
}
