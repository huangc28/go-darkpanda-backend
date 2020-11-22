package chattests

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/huangc28/go-darkpanda-backend/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/teris-io/shortid"
)

type EmitMessageTestSuite struct {
	suite.Suite
}

func (suite *EmitMessageTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
}

func (suite *EmitMessageTestSuite) TestEmitMessageSuccess() {
	// seed a male user
	// the male user send a message to specified channel
	maleUserParams, err := util.GenTestUserParams()
	maleUserParams.Gender = models.GenderMale
	q := models.New(db.GetDB())
	ctx := context.Background()
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	sid, _ := shortid.Generate()
	pubnubChannelID := fmt.Sprintf("private_chat:%s", sid)

	params := &url.Values{}
	params.Add("content", "hello world")
	params.Add("channel_id", pubnubChannelID)

	if err != nil {
		suite.T().Fatal(err)
	}

	req, err := util.ComposeTestRequest(
		"POST",
		"/v1/chat/emit-text-message",
		params,
		util.CreateJwtHeaderMap(
			maleUser.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	ctrl := gomock.NewController(suite.T())
	mChatDao := mock.NewMockChatDaoer(ctrl)
	mChatDao.
		EXPECT().
		GetChatRoomByChannelUUID(
			gomock.Any(),
			gomock.Eq("expired_at"),
			gomock.Eq("message_count"),
		).
		Return(&models.Chatroom{
			MessageCount: sql.NullInt32{
				Int32: 101,
				Valid: true,
			},
			ExpiredAt: time.Now().Add(time.Minute * 27),
		}, nil)

	handlers := chat.ChatHandlers{
		ChatDao: mChatDao,
	}

	c.Request = req
	handlers.EmitTextMessage(c)
	apperr.HandleError()(c)

	// assertions
	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, w.Code)
}

func TestEmitMessageTestSuite(t *testing.T) {
	suite.Run(t, new(EmitMessageTestSuite))
}
