package chattests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GetChatroomMessagesTestSuite struct {
	suite.Suite
	depCon container.Container
}

func (suite *GetChatroomMessagesTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background()).Run(func() {
		deps.Get().Run()
		suite.depCon = deps.Get().Container
	})
}

// GetChatroomMessagesSuccess we are seeding real firestore messages to testify
// the retrieval of chat messages.
func (suite *GetChatroomMessagesTestSuite) TestGetChatroomMessagesSuccess() {
	// Emit 20 test messages.
	ctx := context.Background()

	for i := 1; i <= 20; i++ {
		_, err := darkfirestore.
			Get().
			Client.
			Collection(darkfirestore.PrivateChatsCollectionName).
			Doc("test_chat").
			Collection(darkfirestore.MessageSubCollectionName).
			Doc(fmt.Sprintf("message#%d", i)).
			Set(ctx, darkfirestore.ChatMessage{
				Content:   fmt.Sprintf("message #%d", i),
				From:      "userA",
				To:        "userB",
				CreatedAt: time.Now(),
			})

		if err != nil {
			suite.T().Fatal(err)
		}
	}

	q := models.New(db.GetDB())
	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = models.GenderMale
	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// Now we try to emit an API to retrieve historial messages.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{
		Key:   "channel_uuid",
		Value: "test_chat",
	})

	// Fetch first page.
	req, err := util.ComposeTestRequest(
		"GET",
		fmt.Sprintf("/v1/chat/%s/messages?perpage=5&page=1", "test_chat"),
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
	chat.GetHistoricalMessages(c)
	apperr.HandleError()(c)

	// page1 := chat.TransformedGetHistoricalMessages{}
	type Page1Messages []map[string]string
	page1 := make(map[string]Page1Messages)
	if err := json.Unmarshal(w.Body.Bytes(), &page1); err != nil {
		suite.T().Fatal(err)
	}

	// Assert that message is retrieve from a timestamp descending order fashion.
	assert := assert.New(suite.T())
	assert.Equal(len(page1["messages"]), 5)
	for i := 0; i < 5; i++ {
		tag := 20 - i
		assert.Equal(page1["messages"][i]["content"], fmt.Sprintf("message #%d", tag))
	}

	// Fetch second page.
	w.Body = bytes.NewBuffer([]byte{})
	req2, err := util.ComposeTestRequest(
		"GET",
		fmt.Sprintf("/v1/chat/%s/messages?perpage=5&page=2", "test_chat"),
		&url.Values{},
		util.CreateJwtHeaderMap(
			femaleUser.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	c.Request = req2
	chat.GetHistoricalMessages(c)
	apperr.HandleError()(c)

	type Page2Messages []map[string]string
	page2 := make(map[string]Page2Messages)
	if err := json.Unmarshal(w.Body.Bytes(), &page2); err != nil {
		suite.T().Fatal(err)
	}

	for i := 0; i < 5; i++ {
		tag := 15 - i
		assert.Equal(page2["messages"][i]["content"], fmt.Sprintf("message #%d", tag))
	}
}

func TestGetChatroomMessagesTestSuite(t *testing.T) {
	suite.Run(t, new(GetChatroomMessagesTestSuite))
}
