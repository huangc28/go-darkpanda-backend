package inquirytests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InquiryTestSuite struct {
	suite.Suite
	sendRequest      util.SendRequest
	newUserParams    *models.CreateUserParams
	newInquiryParams *models.CreateInquiryParams
}

func (suite *InquiryTestSuite) SetupSuite() {
	manager.NewDefaultManager()
	suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
}

func (suite *InquiryTestSuite) BeforeTest(suiteName, testName string) {
	// generate new user params before test
	newUserParams, err := util.GenTestUserParams(context.Background())

	if err != nil {
		suite.T().Fatal(err)
	}

	suite.newUserParams = newUserParams
}

func (suite *InquiryTestSuite) TestEmitInquirySuccess() {
	ctx := context.Background()
	newUserParams, err := util.GenTestUserParams(ctx)

	if err != nil {
		suite.T().Fatal(err)
	}

	newUserParams.Gender = models.GenderMale
	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}
	body := struct {
		Budget      float32 `json:"budget"`
		ServiceType string  `json:"service_type"`
	}{
		100.10,
		string(models.ServiceTypeSex),
	}

	if err != nil {
		suite.T().Fatal(err)
	}

	jwt, err := jwtactor.CreateToken(
		newUser.Uuid,
		config.GetAppConf().JwtSecret,
	)
	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", jwt)
	resp, _ := suite.sendRequest(
		"POST",
		"/v1/inquiries",
		body,
		header,
	)

	// ------------------- assert test case -------------------
	respBody := struct {
		Uuid          string               `json:"uuid"`
		Budget        float64              `json:"budget"`
		ServiceType   models.ServiceType   `json:"service_type"`
		InquiryStatus models.InquiryStatus `json:"inquiry_status"`
		CreatedAt     time.Time            `json:"created_at"`
	}{}
	dec := json.NewDecoder(resp.Result().Body)
	dec.Decode(&respBody)

	assert.NotEmpty(suite.T(), respBody.Uuid)
	assert.Equal(suite.T(), respBody.Budget, 100.10)
	assert.Equal(suite.T(), respBody.ServiceType, models.ServiceTypeSex)
	assert.Equal(suite.T(), respBody.InquiryStatus, models.InquiryStatusInquiring)
}

func (suite *InquiryTestSuite) TestCancelInquirySuccess() {
	ctx := context.Background()
	newUserParams := suite.newUserParams
	newUserParams.Gender = models.GenderMale
	q := models.New(db.GetDB())
	usr, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	newInquiryParams, err := util.GenTestInquiryParams(usr.ID)

	if err != nil {
		suite.T().Fatal(err)
	}

	newInquiryParams.InquiryStatus = models.InquiryStatusInquiring
	newInquiry, err := q.CreateInquiry(ctx, *newInquiryParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- request API -------------------
	header := util.CreateJwtHeaderMap(usr.Uuid, config.GetAppConf().JwtSecret)

	resp, err := suite.sendRequest(
		"PATCH",
		fmt.Sprintf("/v1/inquiries/%s/cancel", newInquiry.Uuid),
		struct{}{},
		header,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	respBody := struct {
		Uuid          string `json:"uuid"`
		InquiryStatus string `json:"inquiry_status"`
		Budget        string `json:"budget"`
	}{}

	dec := json.NewDecoder(resp.Result().Body)
	dec.Decode(&respBody)

	// ------------------- assert test case -------------------
	assert.Equal(suite.T(), http.StatusOK, resp.Result().StatusCode)
	assert.Equal(suite.T(), string(models.InquiryStatusCanceled), respBody.InquiryStatus)

	siq, _ := q.GetInquiryByUuid(ctx, newInquiry.Uuid)

	assert.Equal(suite.T(), models.InquiryStatusCanceled, siq.InquiryStatus)
	assert.NotEmpty(suite.T(), respBody.Budget)
}

func (suite *InquiryTestSuite) TestPickupInquirySuccess() {
	ctx := context.Background()

	// create a female user to pickup the inquiry
	femaleUserParams := suite.newUserParams
	femaleUserParams.Gender = models.GenderFemale
	q := models.New(db.GetDB())
	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

	if err != nil {
		suite.T().Fatalf("Failed to create female user %s", err.Error())
	}

	// create a male that hosts the inquiry
	maleUserParams, _ := util.GenTestUserParams(ctx)
	maleUserParams.Gender = models.GenderMale
	maleUser, err := q.CreateUser(ctx, *maleUserParams)

	if err != nil {
		suite.T().Fatalf("Failed to create female user %s", err.Error())
	}

	// create an inquiry
	iqParams, _ := util.GenTestInquiryParams(maleUser.ID)
	iqParams.InquiryStatus = models.InquiryStatusInquiring
	iqParams.ServiceType = models.ServiceTypeSex
	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		suite.T().Fatalf("Failed to create new inquiry %s", err.Error())
	}

	headerMap := util.CreateJwtHeaderMap(femaleUser.Uuid, config.GetAppConf().JwtSecret)
	resp, err := suite.sendRequest(
		"POST",
		fmt.Sprintf("/v1/inquiries/%s/pickup", iq.Uuid),
		struct{}{},
		headerMap,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- Assert test cases -------------------
	assert.Equal(suite.T(), http.StatusOK, resp.Result().StatusCode)

	respBody := inquiry.TransformedService{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(suite.T(), respBody.Uuid)
	assert.Equal(suite.T(), string(models.ServiceStatusUnpaid), respBody.ServiceStatus)
	assert.Equal(suite.T(), string(models.ServiceTypeSex), respBody.ServiceType)

	assert.NotEmpty(suite.T(), respBody.User.Uuid)
	assert.Equal(suite.T(), maleUser.Username, respBody.User.Username)
	assert.Equal(suite.T(), string(maleUser.PremiumType), respBody.User.PremiumType)
}

func TestInquirySuites(t *testing.T) {
	suite.Run(t, new(InquiryTestSuite))
}
