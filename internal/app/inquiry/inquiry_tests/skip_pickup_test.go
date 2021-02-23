package inquirytests

import (
	"context"
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
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/inquiry_tests/helpers"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SkipPickupTestSuite struct {
	suite.Suite
	depCon container.Container
}

func (s *SkipPickupTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			s.depCon = deps.Get().Container
		})
}

func (s *SkipPickupTestSuite) TestSkipPickupSuccess() {
	iqRes := helpers.CreateInquiryStatusUser(
		s.T(),
		helpers.CreateInquiryStatusParam{
			Status: models.InquiryStatusAsking,
		},
	)

	header := util.CreateJwtHeaderMap(
		iqRes.Inquirer.Uuid,
		config.GetAppConf().JwtSecret,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/inquiries/:inquiry/skip",
		&url.Values{},
		header,
	)

	if err != nil {
		s.T().Fatal(err)
	}

	c.Request = req
	c.Params = append(
		c.Params,
		gin.Param{
			Key:   "inquiry_uuid",
			Value: iqRes.Inquiry.Uuid,
		},
	)

	jwtactor.JwtValidator(
		jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		},
	)(c)
	inquiry.SkipPickupHandler(c, s.depCon)
	apperr.HandleError()(c)

	// ------------------- Assertions -------------------
	assert := assert.New(s.T())
	ctx := context.Background()
	// Inquiry record in DB should be `skip`
	var iq models.ServiceInquiry

	db := db.GetDB()
	db.QueryRowx(`
SELECT inquiry_status
FROM service_inquiries
WHERE uuid = $1;
	`, iqRes.Inquiry.Uuid).StructScan(&iq)

	assert.Equal(models.InquiryStatusInquiring, iq.InquiryStatus)
	// Inquiry record in firestore should be `skip`
	fsClient := darkfirestore.Get().Client
	fsResp, err := fsClient.
		Collection("inquiries").
		Doc(iqRes.Inquiry.Uuid).
		Get(ctx)

	if err != nil {
		s.T().Fatal(err)
	}

	assert.Equal(
		string(models.InquiryStatusInquiring),
		fsResp.Data()["status"],
	)
}

func TestSkipPickupTestSuite(t *testing.T) {
	suite.Run(t, new(SkipPickupTestSuite))
}
