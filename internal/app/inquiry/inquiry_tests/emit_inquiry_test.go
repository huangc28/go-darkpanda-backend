package inquirytests

import (
	"context"
	"fmt"
	"log"
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
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type EmitInquiryTestSuite struct {
	suite.Suite
	SendUrlEncodedRequest util.SendUrlEncodedRequest
}

func (suite *EmitInquiryTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).Run(func() {
		deps.Get().Run()
	})

	suite.SendUrlEncodedRequest = util.SendUrlEncodedRequestToApp(app.StartApp(gin.Default()))
}

func (suite *EmitInquiryTestSuite) TestEmitInquirySuccess() {
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

	header := util.CreateJwtHeaderMap(
		newUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	values := url.Values{}
	values.Set("budget", "100.10")
	values.Set("service_type", string(models.ServiceTypeSex))
	values.Set("appointment_time", time.Date(2020, 1, 23, 2, 50, 00, 00, time.UTC).Format(time.RFC3339))
	values.Set("service_duration", fmt.Sprintf("%d", 30))

	// Request API

	//jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
	//Secret: config.GetAppConf().JwtSecret,
	//}),

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/inquiries",
		&values,
		header,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req

	jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	})(c)
	inquiry.EmitInquiryHandler(c)
	apperr.HandleError()(c)

	log.Printf("DEBUG resp %v", w.Body.String())

	// ------------------- assert test case -------------------
	//respBody := struct {
	//Uuid          string               `json:"uuid"`
	//Budget        string               `json:"budget"`
	//ServiceType   models.ServiceType   `json:"service_type"`
	//ChannelID     string               `json:"channel_id"`
	//InquiryStatus models.InquiryStatus `json:"inquiry_status"`
	//CreatedAt     time.Time            `json:"created_at"`
	//}{}

	//dec := json.NewDecoder(resp.Result().Body)
	//dec.Decode(&respBody)
	//assert := assert.New(suite.T())

	//assert.NotEmpty(respBody.Uuid)
	//assert.Equal(respBody.Budget, "100.10")
	//assert.Equal(respBody.ServiceType, models.ServiceTypeSex)
	//assert.Equal(respBody.InquiryStatus, models.InquiryStatusInquiring)
	//assert.NotEmpty(respBody.ChannelID)
}

func TestEmitInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(EmitInquiryTestSuite))
}
