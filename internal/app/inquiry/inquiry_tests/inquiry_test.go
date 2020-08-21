package inquirytests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InquiryTestSuite struct {
	suite.Suite
	sendRequest util.SendRequest
}

func (suite *InquiryTestSuite) SetupSuite() {
	manager.NewDefaultManager()
	suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
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

func TestInquirySuites(t *testing.T) {
	suite.Run(t, new(InquiryTestSuite))
}
