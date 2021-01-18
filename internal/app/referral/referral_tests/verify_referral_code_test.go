package referral_tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
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
	"github.com/huangc28/go-darkpanda-backend/internal/app/referral"
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

func (suite *VerifyReferralCodeTestSuite) invitorInviteeProvider(ctx context.Context) []models.User {
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

	return mans
}

func (suite *VerifyReferralCodeTestSuite) TestVerifyInvitorCodeExpired() {
}

func (suite *VerifyReferralCodeTestSuite) TestVerifyInvitorCodeInvalId() {
	ctx := context.Background()
	mans := suite.invitorInviteeProvider(ctx)

	invitor := mans[0]
	invitee := mans[1]

	q := models.New(db.GetDB())
	refCode, err := q.CreateRefcode(
		ctx,
		models.CreateRefcodeParams{
			InvitorID: int32(invitor.ID),
			InviteeID: sql.NullInt32{
				Valid: true,
				Int32: int32(invitee.ID),
			},
			RefCode:     util.GenRandStringRune(10),
			RefCodeType: models.RefCodeTypeInvitor,
		},
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	// 3rd invitee is trying to use the occupied referral code...
	otherManParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatal(err)
	}

	otherManParams.Gender = models.GenderMale
	otherMan, err := q.CreateUser(ctx, *otherManParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	log.Printf("DEBUG otherMan %v", otherMan)

	// Request the API.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := &url.Values{}
	params.Add("invitee_uuid", otherMan.Uuid)
	params.Add("referral_code", refCode.RefCode)
	headers := make(map[string]string)

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/referral/verify",
		params,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	referral.HandleVerifyReferralCode(c, suite.depCon)
	apperr.HandleError()(c)

	suite.assert.Equal(http.StatusBadRequest, w.Result().StatusCode)

	body := struct {
		ErrCode string `json:"err_code"`
	}{}

	json.Unmarshal(w.Body.Bytes(), &body)
	suite.assert.Equal(apperr.ReferralCodeIsOccupied, body.ErrCode)
}

func (suite *VerifyReferralCodeTestSuite) TestVerifyInvitorReferralCodeSuccess() {
	ctx := context.Background()
	mans := suite.invitorInviteeProvider(ctx)

	invitor := mans[0]
	invitee := mans[1]

	// Seed sample referral code.
	q := models.New(db.GetDB())
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
		"/v1/referral/verify",
		params,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	referral.HandleVerifyReferralCode(c, suite.depCon)
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
