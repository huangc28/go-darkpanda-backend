package inquirytests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PickupInquiryTestSuite struct {
	depCon container.Container
	suite.Suite
}

func (suite *PickupInquiryTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
	deps.Get().Run()
	suite.depCon = deps.Get().Container
}

func (suite *PickupInquiryTestSuite) TestPickupInquirySuccess() {
	ctx := context.Background()

	// create a female user to pickup the inquiry
	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = models.GenderFemale
	q := models.New(db.GetDB())
	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

	if err != nil {
		suite.T().Fatalf("Failed to create female user %s", err.Error())
	}

	// create a male that hosts the inquiry
	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = models.GenderMale
	maleUser, err := q.CreateUser(ctx, *maleUserParams)

	if err != nil {
		suite.T().Fatalf("Failed to create male user %s", err.Error())
	}

	// create an inquiry
	iqParams, _ := util.GenTestInquiryParams(maleUser.ID)
	iqParams.PickerID = sql.NullInt32{
		Valid: true,
		Int32: int32(femaleUser.ID),
	}
	iqParams.InquiryStatus = models.InquiryStatusInquiring
	iqParams.ServiceType = models.ServiceTypeSex
	iqParams.ExpiredAt = sql.NullTime{
		Time:  time.Now().Add(time.Minute * 27),
		Valid: true,
	}
	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		suite.T().Fatalf("Failed to create new inquiry %s", err.Error())
	}

	// male user joins the lobby
	var df darkfirestore.DarkFireStorer
	suite.depCon.Make(&df)

	df.CreateInquiringUser(
		ctx,
		darkfirestore.CreateInquiringUserParams{
			InquiryUUID:   iq.Uuid,
			InquiryStatus: string(models.InquiryStatusInquiring),
		},
	)

	if err != nil {
		suite.T().Fatalf("Failed to join lobby %s", err.Error())
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := util.ComposeTestRequest(
		"POST",
		fmt.Sprintf("/v1/inquiries/%s/pickup", iq.Uuid),
		&url.Values{},
		util.CreateJwtHeaderMap(femaleUser.Uuid, config.GetAppConf().JwtSecret),
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	c.Params = append(c.Params, gin.Param{
		Key:   "inquiry_uuid",
		Value: iq.Uuid,
	})

	inquiry.PickupInquiryHandler(c, suite.depCon)
	apperr.HandleError()(c)

	respStruct := struct {
		ServiceType   string `json:"service_type"`
		InquiryUUID   string `json:"inquiry_uuid"`
		InquiryStatus string `json:"inquiry_status"`
		ExpiredAt     string `json:"expired_at"`
		CreatedAt     string `json:"created_at"`
	}{}

	json.Unmarshal(w.Body.Bytes(), &respStruct)

	// ------------------- Assert test cases -------------------
	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, w.Result().StatusCode)

	// Assert that the lobby user status in DB has changed to `waiting`
	var si models.ServiceInquiry
	db := db.GetDB()
	if err := db.QueryRowx(
		`
	SELECT
		uuid,
		inquiry_status,
		picker_id
	FROM
		service_inquiries
	WHERE
		id = $1;
	`,
		iq.ID,
	).StructScan(&si); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(
		models.InquiryStatusAsking,
		si.InquiryStatus,
	)

	assert.Equal(
		int64(si.PickerID.Int32),
		femaleUser.ID,
	)

	// Assert that the user status in firestore has changed to `waiting`
	dfClient := df.GetClient()
	dfResp, err := dfClient.
		Collection("inquiries").
		Doc(respStruct.InquiryUUID).
		Get(ctx)

	if err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(
		dfResp.Data()["status"],
		"asking",
	)
}

func TestPickupInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(PickupInquiryTestSuite))
}
