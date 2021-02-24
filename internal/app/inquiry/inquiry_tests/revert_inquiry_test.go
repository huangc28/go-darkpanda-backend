package inquirytests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/inquiry_tests/helpers"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
)

type RevertRevertInquiryTestSuite struct {
	suite.Suite
	container container.Container
}

func (suite *RevertRevertInquiryTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			suite.container = deps.Get().Container
		})
}

func (suite *RevertRevertInquiryTestSuite) TestRevertInquirySuccess() {
	// Seed female (service provider) and male user (inquirer)
	iqResp := helpers.CreateInquiryStatusUser(
		suite.T(),
		helpers.CreateInquiryStatusParam{
			Status: models.InquiryStatusChatting,
		},
	)

	ctx := context.Background()
	q := models.New(db.GetDB())
	chatParams, _ := util.GenTestChat(iqResp.Inquiry.ID)
	chatroom, err := q.CreateChatroom(ctx, *chatParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	chatDao := chat.NewChatDao(db.GetDB())
	if err := chatDao.JoinChat(
		chatroom.ID,
		iqResp.Inquirer.ID,
		iqResp.Picker.ID,
	); err != nil {
		suite.T().Fatal(err)
	}

	// Request API
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := util.ComposeTestRequest(
		"PATCH",
		fmt.Sprintf("inquiries/%s/revert-chat", iqResp.Inquiry.Uuid),
		&url.Values{},
		util.CreateJwtHeaderMap(
			iqResp.Picker.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	c.Request = req

	c.Set("uuid", iqResp.Inquiry.Uuid)
	c.Set("inquiry", &iqResp.Inquiry)
	inquiry.RevertChatHandler(c, suite.container)
	apperr.HandleError()(c)

	log.Printf(
		"DEBUG inquiries/%s/revert-chat response %v",
		iqResp.Inquiry.Uuid,
		w.Body.String(),
	)

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
	`, chatroom.ID, iqResp.Picker.ID).Scan(&femaleUserNotExists); err != nil {
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
	`, iqResp.Inquiry.ID).Scan(&isStatusInquiring); err != nil {
		suite.T().Fatal(err)
	}

	assert.True(isStatusInquiring)

	// Check firestore inquiry status has been changed to inquiring
	dfClient := darkfirestore.Get().Client
	dfResp, err := dfClient.
		Collection("inquiries").
		Doc(iqResp.Inquiry.Uuid).
		Get(ctx)

	if err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(
		string(models.InquiryStatusInquiring),
		dfResp.Data()["status"],
	)
}

func TestRevertInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(RevertRevertInquiryTestSuite))
}
