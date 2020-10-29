package inquirytests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RevertRevertInquiryTestSuite struct {
	suite.Suite
}

func (suite *RevertRevertInquiryTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
}

func (suite *RevertRevertInquiryTestSuite) TestRevertInquiry() {
	// Seed female (service provider) and male user (inquirer)
	ctx := context.Background()
	q := models.New(db.GetDB())

	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = models.GenderMale
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = models.GenderFemale
	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	// Seed inquiry with chatting status.
	testIqParams, _ := util.GenTestInquiryParams(maleUser.ID)
	testIqParams.ExpiredAt = sql.NullTime{
		Time:  time.Now().Add(time.Minute * 27),
		Valid: true,
	}
	testIqParams.InquiryStatus = models.InquiryStatusChatting
	iq, err := q.CreateInquiry(ctx, *testIqParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	chatParams, _ := util.GenTestChat(iq.ID)
	chatroom, err := q.CreateChatroom(ctx, *chatParams)
	log.Printf("DEBUG chatroom ~ %v ", chatroom.ChannelUuid)
	if err != nil {
		suite.T().Fatal(err)
	}

	chatDao := chat.NewChatDao(db.GetDB())
	if err := chatDao.JoinChat(chatroom.ID, maleUser.ID, femaleUser.ID); err != nil {
		suite.T().Fatal(err)
	}

	// Request API
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := util.ComposeTestRequest(
		"PATCH",
		fmt.Sprintf("inquiries/%s/revert-chat", iq.Uuid),
		&url.Values{},
		util.CreateJwtHeaderMap(
			femaleUser.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req
	handlers := inquiry.InquiryHandlers{
		ChatDao: chat.NewChatDao(db.GetDB()),
		UserDao: user.NewUserDAO(db.GetDB()),
	}

	c.Set("uuid", femaleUser.Uuid)
	c.Set("inquiry", &iq)
	handlers.RevertChat(c)
	apperr.HandleError()(c)

	// Assertions
	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, w.Code)
	dbi := db.GetDB()
	var (
		femaleUserNotExists bool
		chatroomNotExists   bool
		isStatusInquiring   bool
	)

	// Assert female user is not in chat
	if err := dbi.DB.QueryRow(`
SELECT EXISTS (
	SELECT 1 FROM chatroom_users
	WHERE deleted_at IS NOT NULL
	AND chatroom_id = $1
	AND user_id = $2
) as exists;
	`, chatroom.ID, femaleUser.ID).Scan(&femaleUserNotExists); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(femaleUserNotExists)

	// Assert chatroom has been removed
	if err := dbi.DB.QueryRow(`
SELECT EXISTS (
	SELECT 1 FROM chatrooms
	WHERE deleted_at IS NOT NULL
	AND id = $1
) as exists;
	`, chatroom.ID).Scan(&chatroomNotExists); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(chatroomNotExists)
	// Check inquiry status is changed to inquiring
	if err := dbi.DB.QueryRow(`
SELECT EXISTS (
	SELECT 1 FROM service_inquiries
	WHERE id = $1
	AND inquiry_status = 'inquiring'
) as exists;
	`, iq.ID).Scan(&isStatusInquiring); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(isStatusInquiring)

	trf := inquiry.TransformedRevertChatting{}
	if err := json.Unmarshal(w.Body.Bytes(), &trf); err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(trf.LobbyChannelID)
}

func TestRevertInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(RevertRevertInquiryTestSuite))
}
