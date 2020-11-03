package inquirytests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PickupInquiryTestSuite struct {
	suite.Suite
	SendUrlEncodedRequest util.SendUrlEncodedRequest
}

func (suite *PickupInquiryTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
	deps.Get().Run()

	suite.SendUrlEncodedRequest = util.SendUrlEncodedRequestToApp(app.StartApp(gin.Default()))
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
	lobbySrv := inquiry.LobbyServices{
		LobbyDao: &inquiry.LobbyDao{
			DB: db.GetDB(),
		},
	}

	_, err = lobbySrv.JoinLobby(iq.ID)

	if err != nil {
		suite.T().Fatalf("Failed to join lobby %s", err.Error())
	}

	headerMap := util.CreateJwtHeaderMap(femaleUser.Uuid, config.GetAppConf().JwtSecret)
	resp, err := suite.SendUrlEncodedRequest(
		"POST",
		fmt.Sprintf("/v1/inquiries/%s/pickup", iq.Uuid),
		&url.Values{},
		headerMap,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- Assert test cases -------------------
	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, resp.Result().StatusCode)

	// assert that the inquiry has been removed from lobby (soft deleted).
	db := db.GetDB()
	var removedUserExists bool
	if err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM lobby_users
	WHERE inquiry_id = $1
	AND deleted_at IS NOT NULL
) AS exists;
	`, iq.ID).Scan(&removedUserExists); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(removedUserExists)

	// assert that both male and female user are in the chatroom already.
	var (
		maleExistsInChat     bool
		femaleExistsInChat   bool
		pickerIDIsFemaleUser bool
	)
	existenceQuery := `
SELECT EXISTS(
	SELECT 1 FROM chatroom_users
	WHERE user_id = $1
) AS exists;
	`
	if err := db.QueryRow(existenceQuery, maleUser.ID).Scan(&maleExistsInChat); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(maleExistsInChat)

	if err := db.QueryRow(existenceQuery, femaleUser.ID).Scan(&femaleExistsInChat); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(femaleExistsInChat)

	// assert the value of picker_id on inquiry is female user
	if err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM service_inquiries
	WHERE id = $1 
	AND picker_id = $2
) AS exists;
	`, iq.ID, femaleUser.ID).Scan(&pickerIDIsFemaleUser); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(pickerIDIsFemaleUser)

	respBody := inquiry.TransformedPickupInquiry{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(respBody.ChannelUUID)
	assert.Equal(string(models.ServiceTypeSex), respBody.ServiceType)
	assert.Equal(string(models.InquiryStatusChatting), respBody.InquiryStatus)

	assert.NotEmpty(respBody.InquirerUUID)
	assert.Equal(maleUser.Username, respBody.Username)
}

func TestPickupInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(PickupInquiryTestSuite))
}
