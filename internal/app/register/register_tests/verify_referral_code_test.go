package register_tests

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/register"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VerifyReferralCodeTestSuite struct {
	suite.Suite
	depCon container.Container
	assert *assert.Assertions
}

func (suite *VerifyReferralCodeTestSuite) SetupSuite() {
	suite.assert = assert.New(suite.T())

	manager.
		NewDefaultManager(context.Background()).Run(func() {
		deps.Get().Run()
		suite.depCon = deps.Get().Container
	})
}

func (suite *VerifyReferralCodeTestSuite) TestVerifyInvitorReferralCodeSuccess() {
	ctx := context.Background()
	q := models.New(db.GetDB())

	mans := make([]models.User, 0)

	// creates 2 man, one invitor one invitee.
	for i := 0; i < 2; i++ {
		manParams, err := util.GenTestUserParams()

		if err != nil {
			suite.T().Fatal(err)
		}

		manParams.Gender = models.GenderMale
		man, err := q.CreateUser(ctx, *manParams)

		if err != nil {
			suite.T().Fatal(err)
		}

		mans = append(mans, man)
	}

	invitor := mans[0]
	invitee := mans[1]

	// Seed sample referral code.
	refCode, err := q.CreateRefcode(
		ctx,
		models.CreateRefcodeParams{
			InvitorID: int32(invitor.ID),
			InviteeID: sql.NullInt32{
				Valid: false,
			},
			RefCode:     util.GenRandStringRune(10),
			RefCodeType: models.RefCodeTypeInvitor,
		},
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	// Request the API.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := &url.Values{}
	params.Add("invitee_uuid", invitee.Uuid)
	params.Add("referral_code", refCode.RefCode)

	headers := make(map[string]string)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/register/verify-referral-code",
		params,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	register.HandleVerifyReferralCode(c, suite.depCon)
	apperr.HandleError()(c)

	// Assert that the response is OK.
	suite.Equal(http.StatusOK, w.Code)

	// Assert that the refcode is occupied
	db := db.GetDB()
	var inviteeID int64
	db.QueryRow(`
SELECT invitee_id
FROM user_refcodes
WHERE id = $1;
	`, refCode.ID).Scan(&inviteeID)
	suite.Equal(inviteeID, invitee.ID)
}

func TestVerifyReferralCodeTestSuite(t *testing.T) {
	suite.Run(t, new(VerifyReferralCodeTestSuite))
}
