package usertests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	sendRequest util.SendRequest
}

func (suite *UserAPITestsSuite) SetupSuite() {
	manager.NewDefaultManager()
	suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
}

func (suite *UserAPITestsSuite) TestGetMaleUserInfo() {
	// create a male user that has no related active inquiry
	ctx := context.Background()
	newUserParams, err := util.GenTestUserParams(ctx)
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

	resp, _ := suite.sendRequest("POST", "/v1/users/me", struct{}{}, header)

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
	newUserParams, err := util.GenTestUserParams(ctx)
	newUserParams.Gender = models.GenderMale

	if err != nil {
		suite.T().Fatal(err)
	}

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
	resp, _ := suite.sendRequest("POST", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases -------------------
	respStruct := &user.TransformUserWithInquiryData{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), len(respStruct.Inquiries), 1)
}

func (suite *UserAPITestsSuite) TestPutUserInfoSuccess() {
	// ------------------- create male user -------------------
	ctx := context.Background()
	maleUserParams, _ := util.GenTestUserParams(ctx)
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

func TestUserAPISuite(t *testing.T) {
	suite.Run(t, new(UserAPITestsSuite))
}
