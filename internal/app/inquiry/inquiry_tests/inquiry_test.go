package inquirytests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/inquiry_tests/helpers"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/shopspring/decimal"
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
	manager.NewDefaultManager(context.Background())
	//suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
}

func (suite *InquiryTestSuite) BeforeTest(suiteName, testName string) {
	// generate new user params before test
	newUserParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatal(err)
	}

	suite.newUserParams = newUserParams
}

func (suite *InquiryTestSuite) TestCancelInquirySuccess() {
	iqResp := helpers.CreateInquiryStatusUser(
		suite.T(),
		helpers.CreateInquiryStatusParam{
			Status: models.InquiryStatusInquiring,
		},
	)

	// ------------------- request API -------------------
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeTestRequest(
		"PATCH",
		fmt.Sprintf(
			"/v1/inquiries/%s/cancel",
			iqResp.Inquiry.Uuid,
		),
		&url.Values{},
		util.CreateJwtHeaderMap(
			iqResp.Inquirer.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	c.Params = append(
		c.Params,
		gin.Param{
			Key:   "inquiry_uuid",
			Value: iqResp.Inquiry.Uuid,
		},
	)

	jwtactor.JwtValidator(
		jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		},
	)(c)
	inquiry.ValidateInqiuryURIParams()(c)
	inquiry.ValidateBeforeAlterInquiryStatus(inquiry.Cancel)(c)
	inquiry.CancelInquiryHandler(c)
	apperr.HandleError()(c)

	// ------------------- Assertions -------------------
	respBody := struct {
		Uuid          string  `json:"inquiry_uuid"`
		InquiryStatus string  `json:"inquiry_status"`
		Budget        float64 `json:"budget"`
		ServiceType   string  `json:"service_type"`
	}{}

	if err = json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- Assert test case -------------------
	assert := assert.New(suite.T())
	ctx := context.Background()

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(string(models.InquiryStatusCanceled), respBody.InquiryStatus)

	dfClient := darkfirestore.Get().Client
	iqDoc, err := dfClient.
		Collection("inquiries").
		Doc(iqResp.Inquiry.Uuid).
		Get(ctx)

	if err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(
		string(models.InquiryStatusCanceled),
		iqDoc.Data()["status"],
	)
}

func (suite *InquiryTestSuite) TestGirlApproveInquirySuccess() {
	// Create male / female user
	ctx := context.Background()
	maleUserParams := suite.newUserParams
	maleUserParams.Gender = models.GenderMale

	q := models.New(db.GetDB())
	maleUser, _ := q.CreateUser(ctx, *maleUserParams)

	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = models.GenderFemale
	femaleUser, _ := q.CreateUser(ctx, *femaleUserParams)

	iqParams, _ := util.GenTestInquiryParams(maleUser.ID)

	iqParams.InquiryStatus = models.InquiryStatusChatting
	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	headers := util.CreateJwtHeaderMap(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	body := struct {
		Price           float64   `json:"price"`
		Duration        int       `json:"duration"`
		AppointmentTime time.Time `json:"appointment_time"`
		Lng             float64   `json:"lng"`
		Lat             float64   `json:"lat"`
	}{
		3500,
		120,
		time.Now().Add(time.Hour * 48),
		25.0806874,
		121.5495119,
	}

	resp, err := suite.sendRequest(
		"POST",
		fmt.Sprintf("/v1/inquiries/%s/girl-approve", iq.Uuid),
		&body,
		headers,
	)

	if err != nil {
		log.Fatal(err)
	}

	// ------------------- assert test cases -------------------
	respBody := inquiry.TransformedGirlApproveInquiry{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(&respBody); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, resp.Result().StatusCode)
	assert.Equal(string(models.InquiryStatusWaitForInquirerApprove), respBody.InquiryStatus)

	assert.Equal("3500.00", respBody.Price)
	assert.Equal("25.0806874", respBody.Lng)
	assert.Equal("121.5495119", respBody.Lat)
}

func (suite *InquiryTestSuite) TestManBooksInquirySuccess() {
	// ------------------- create test data -------------------
	// Create male / female user
	ctx := context.Background()
	maleUserParams := suite.newUserParams
	maleUserParams.Gender = models.GenderMale

	q := models.New(db.GetDB())
	maleUser, _ := q.CreateUser(ctx, *maleUserParams)

	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = models.GenderFemale
	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	iqParams, _ := util.GenTestInquiryParams(maleUser.ID)

	iqParams.InquiryStatus = models.InquiryStatusWaitForInquirerApprove
	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- request API -------------------
	body := struct {
		Price               float64   `json:"price"`
		Duration            int       `json:"duration"`
		AppointmentTime     time.Time `json:"appointment_time"`
		Lng                 float64   `json:"lng"`
		Lat                 float64   `json:"lat"`
		ServiceType         string    `json:"service_type"`
		ServiceProviderUuid string    `json:"service_provider_uuid"`
	}{
		3600,
		180,
		time.Now().Add(time.Hour * 24 * 4),
		25.0806874,
		121.5495119,
		string(models.ServiceTypeSex),
		femaleUser.Uuid,
	}

	headers := util.CreateJwtHeaderMap(
		maleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, err := suite.sendRequest(
		"POST",
		fmt.Sprintf("/v1/inquiries/%s/book", iq.Uuid),
		&body,
		headers,
	)

	if err != nil {
		suite.T().Fatalf("request failed: %s", err.Error())
	}

	if resp.Result().StatusCode != http.StatusOK {
		bbody, _ := ioutil.ReadAll(resp.Result().Body)
		suite.T().Fatalf("request failed %s", string(bbody))
	}

	// ------------------- assert test case -------------------
	assert := assert.New(suite.T())
	respBody := &inquiry.TransformedBookedService{}

	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(femaleUser.Username, respBody.ServiceProvider.Username)
	assert.Equal(femaleUser.Uuid, respBody.ServiceProvider.Uuid)

	assert.Equal(respBody.Lat, fmt.Sprintf("%s0", decimal.NewFromFloat(body.Lat).String()))
	assert.Equal(respBody.Lng, fmt.Sprintf("%s0", decimal.NewFromFloat(body.Lng).String()))
}

// GetInquiriesSuite test cases when retrieving inquiries.
type GetInquiriesSuite struct {
	suite.Suite
	sendRequest util.SendRequest
}

func (suite *GetInquiriesSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
		})

	suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
}

func (s *GetInquiriesSuite) TestGetInquiriesSuccess() {
	// create 5 male users with phone verified.
	ctx := context.Background()
	q := models.New(db.GetDB())
	maleUsers := make([]*models.User, 5)

	for i := range maleUsers {
		p, err := util.GenTestUserParams()

		if err != nil {
			s.T().Fatal(err)
		}

		p.Gender = models.GenderMale
		p.PhoneVerified = true

		maleUser, err := q.CreateUser(ctx, *p)

		if err != nil {
			s.T().Fatal("failed to create user", err)
		}

		maleUsers[i] = &maleUser
	}

	// create an female user
	femaleParams, _ := util.GenTestUserParams()
	femaleParams.Gender = models.GenderFemale
	femaleParams.Username = "girlP"
	femaleUser, err := q.CreateUser(ctx, *femaleParams)

	if err != nil {
		s.T().Fatal("failed to create female user", err)
	}

	// each male user emits an inquiry.
	inquiries := make([]models.ServiceInquiry, 0)
	for _, maleUser := range maleUsers {
		iqParams, err := util.GenTestInquiryParams(maleUser.ID)
		iqParams.InquiryStatus = models.InquiryStatusInquiring

		if err != nil {
			s.T().Fatal(err)
		}

		iqParams.Price = sql.NullString{
			String: "1.1",
			Valid:  true,
		}
		iqParams.ExpiredAt = sql.NullTime{
			Valid: true,
			Time:  time.Now().Add(time.Minute * 27),
		}

		inquiry, err := q.CreateInquiry(ctx, *iqParams)

		if err != nil {
			s.T().Fatal(err)
		}

		inquiries = append(inquiries, inquiry)
	}

	// female user searches for active inquiries...
	headers := util.CreateJwtHeaderMap(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, _ := s.sendRequest(
		"GET",
		"/v1/inquiries",
		&struct{}{},
		headers,
	)

	// ------------------- assert test case -------------------
	respBody := &inquiry.TransformedInquiries{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respBody); err != nil {
		s.T().Fatal(err)
	}

	// pick an element to assert value matches
	assert := assert.New(s.T())

	assert.Equal(5, len(respBody.Inquiries))
	assert.Equal(inquiries[0].Uuid, respBody.Inquiries[0].Uuid)
	assert.Equal(inquiries[1].Uuid, respBody.Inquiries[1].Uuid)

	assert.Equal(maleUsers[0].Uuid, respBody.Inquiries[0].Inquirer.Uuid)
	assert.Equal(maleUsers[1].Uuid, respBody.Inquiries[1].Inquirer.Uuid)
}

func TestInquirySuites(t *testing.T) {
	suite.Run(t, new(InquiryTestSuite))
	suite.Run(t, new(GetInquiriesSuite))
}
