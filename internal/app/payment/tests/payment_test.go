package tests

import (
	"context"
	"database/sql"
	"log"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/payment"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type PaymentTestSuite struct {
	suite.Suite
	depCon container.Container
}

func (suite *PaymentTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			suite.depCon = deps.Get().Container
		})
}

func (suite *PaymentTestSuite) TestPaymentSuccessfullyCharges20Percent() {
	ctx := context.Background()
	q := models.New(db.GetDB())

	resp, err := util.CreateTestService(ctx, q, util.CreateTestServiceHooks{
		InquiryPreCreateHook: func(iq *models.CreateInquiryParams, female *models.User) {
			iq.InquiryType = models.InquiryTypeDirect
			iq.PickerID = sql.NullInt32{
				Valid: true,
				Int32: int32(female.ID),
			}
		},

		ServicePreCreateHook: func(srv *models.CreateServiceParams) {
			srv.Price = sql.NullString{
				Valid:  true,
				String: "1000.00",
			}

			srv.ServiceStatus = models.ServiceStatusUnpaid
		},
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	var ubDao contracts.UserBalancer
	suite.depCon.Make(&ubDao)

	ubDao.CreateOrTopUpBalance(contracts.CreateOrTopUpBalanceParams{
		UserID:      int(resp.Male.ID),
		TopupAmount: 10000,
	})

	// Send requests
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := &url.Values{}
	params.Add("service_uuid", resp.Service.Uuid.String)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/payment",
		params,
		util.CreateJwtHeaderMap(
			resp.Male.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Set("uuid", resp.Male.Uuid)
	c.Request = req
	payment.CreatePayment(c, suite.depCon)
	apperr.HandleError()(c)

	log.Printf("body %v", w.Body.String())
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}
