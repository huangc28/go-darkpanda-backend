package usertests

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"
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

	resp, _ := suite.sendRequest("POST", "/v1/me", struct{}{}, header)

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
	resp, _ := suite.sendRequest("POST", "/v1/me", struct{}{}, header)

	// ------------------- assert test cases -------------------
	respStruct := &user.TransformUserWithInquiryData{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), len(respStruct.Inquiries), 1)
}

func TestUserAPISuite(t *testing.T) {
	suite.Run(t, new(UserAPITestsSuite))
}
