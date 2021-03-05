package service

import (
	"context"
	"log"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/service"
	testhelpers "github.com/huangc28/go-darkpanda-backend/internal/app/test_helpers"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type GetServicesTestSuite struct {
	suite.Suite
	depCon      container.Container
	testHelpers *testhelpers.TestHelpers
}

func (s *GetServicesTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			s.depCon = deps.Get().Container
			s.testHelpers = testhelpers.NewTestHelpers()
		})
}

func (s *GetServicesTestSuite) TestGetCurrentServicesSuccess() {
	// Create stub services
	srv1, err := s.
		testHelpers.
		CreateTestService(testhelpers.CreateTestServiceParams{
			ServiceStatus: models.ServiceStatusUnpaid,
		})

	if err != nil {
		s.T().Fatal(err)
	}

	srv2, err := s.
		testHelpers.
		CreateTestService(
			testhelpers.CreateTestServiceParams{
				ServiceStatus: models.ServiceStatusToBeFulfilled,
				Picker:        srv1.Picker,
			},
		)

	if err != nil {
		s.T().Fatal(err)
	}

	// Create service chatrooms
	var srvDao contracts.ChatServicer
	s.depCon.Make(&srvDao)

	srvDao.CreateAndJoinChatroom(
		srv1.Inquiry.ID,
		srv1.Picker.ID,
		srv1.Inquirer.ID,
	)

	srvDao.CreateAndJoinChatroom(
		srv2.Inquiry.ID,
		srv2.Picker.ID,
		srv2.Inquirer.ID,
	)

	// Retrieve services from the API
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeJsonTestRequest(
		"POST",
		"/v1/services/incoming",
		&url.Values{},
		util.CreateJwtHeaderMap(
			srv1.Picker.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		s.T().Fatal(err)
	}

	c.Request = req

	jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	})(c)

	service.GetListOfCurrentServicesHandler(
		c,
		s.depCon,
	)

	apperr.HandleError()(c)

	log.Printf("DEBUG response %v", w.Body.String())

}

func TestGetServicesTestSuite(t *testing.T) {
	suite.Run(t, new(GetServicesTestSuite))
}
