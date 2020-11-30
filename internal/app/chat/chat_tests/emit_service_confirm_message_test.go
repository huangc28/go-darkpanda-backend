package chattests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/huangc28/go-darkpanda-backend/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EmitServiceConfirmMessageTestSuite struct {
	suite.Suite
}

func (suite *EmitServiceConfirmMessageTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
}

// TestEmitServiceConfirmedMessageSuccess test emiting service confirmed message.
func (suite *EmitServiceConfirmMessageTestSuite) TestEmitServiceConfirmedMessageSuccess() {
	// Create a male user
	ctx := context.Background()
	q := models.New(db.GetDB())
	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = models.GenderMale
	maleUser, err := q.CreateUser(ctx, *maleUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	v := &url.Values{}
	v.Add("price", "10.2")
	v.Add("channel_uuid", "some_channel_uuid")
	v.Add("inquiry_uuid", "some_inquiry_uuid")
	v.Add("service_time", time.Now().Format("2006-01-02T15:04:05Z07:00"))
	v.Add("service_duration", "30")
	v.Add("service_type", string(models.ServiceTypeSex))

	req, err := util.ComposeTestRequest(
		"POST",
		fmt.Sprintf("/v1/chat/emit-service-confirmed-message"),
		v,
		util.CreateJwtHeaderMap(
			maleUser.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	c.Request = req

	if err != nil {
		suite.T().Fatal(err)
	}

	ctrl := gomock.NewController(suite.T())
	mServiceDao := mock.NewMockServiceDAOer(ctrl)

	handlers := chat.ChatHandlers{
		ServiceDao: mServiceDao,
	}

	serviceUUID := uuid.New()
	mServiceDao.EXPECT().
		GetServiceByInquiryUUID(gomock.Eq("some_inquiry_uuid")).
		Return(
			&models.Service{
				Uuid: serviceUUID,
			},
			nil,
		)

	handlers.EmitServiceConfirmedMessage(c)
	apperr.HandleError()(c)

	// Retrieve message from firestore to makesure the correctness of the content.
	resp := struct {
		Message   interface{} `json:"message"`
		MessageID string      `json:"message_id"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		suite.T().Fatal(err)
	}

	_, err = darkfirestore.
		Get().
		Client.
		Collection(darkfirestore.PrivateChatsCollectionName).
		Doc("some_channel_uuid").
		Collection(darkfirestore.MessageSubCollectionName).
		Doc(resp.MessageID).
		Get(ctx)

	if err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Nil(err)

	_, err = darkfirestore.
		Get().
		Client.
		Collection(darkfirestore.PrivateChatsCollectionName).
		Doc("some_channel_uuid").
		Collection(darkfirestore.MessageSubCollectionName).
		Doc(resp.MessageID).
		Delete(ctx)

}

func TestEmitServiceConfirmMessageTestSuite(t *testing.T) {
	suite.Run(t, new(EmitServiceConfirmMessageTestSuite))
}
