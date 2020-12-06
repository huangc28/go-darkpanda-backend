package inquirytests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GetInquirerSuite struct {
	suite.Suite
	depCon container.Container
}

func (s *GetInquirerSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).Run(func() {
		deps.Get().Run()
		s.depCon = deps.Get().Container
	})

}

func (s *GetInquirerSuite) TestGetInquirerInfoSuccess() {
	maleUserParams, err := util.GenTestUserParams()

	if err != nil {
		s.T().Fatal(err)
	}

	maleUserParams.Gender = models.GenderMale
	q := models.New(db.GetDB())
	ctx := context.Background()
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	if err != nil {
		s.T().Fatal(err)
	}

	inquiryParam, err := util.GenTestInquiryParams(maleUser.ID)
	inquiryParam.PickerID = sql.NullInt32{
		Valid: false,
	}
	inquiryParam.Uuid = "example_inquiry_uuid"

	if err != nil {
		s.T().Fatal(err)
	}

	_, err = q.CreateInquiry(ctx, *inquiryParam)

	if err != nil {
		s.T().Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	vals := &url.Values{}

	req, err := util.ComposeTestRequest(
		"GET",
		fmt.Sprintf("/v1/%s/inquirer", "example_inquiry_uuid"),
		vals,
		util.CreateJwtHeaderMap(
			maleUser.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		s.T().Fatal(err)
	}

	c.Params = append(
		c.Params,
		gin.Param{
			Key:   "inquiry_uuid",
			Value: "example_inquiry_uuid",
		},
	)
	c.Request = req
	inquiry.GetInquirerInfo(c, s.depCon)
	apperr.HandleError()(c)

	trfResp := inquiry.TransformGetInquirerInfo{}
	if err := json.Unmarshal(w.Body.Bytes(), &trfResp); err != nil {
		s.T().Fatal(err)
	}

	// Assert that username and uuid are equal to the seeded user
	assert := assert.New(s.T())
	assert.Equal(trfResp.UUID, maleUser.Uuid)
	assert.Equal(trfResp.Username, maleUser.Username)

}

func TestGetInquirerSuite(t *testing.T) {
	suite.Run(t, new(GetInquirerSuite))
}
