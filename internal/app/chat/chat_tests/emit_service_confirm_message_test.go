package chattests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/golobby/container/pkg/container"
	"github.com/google/uuid"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/huangc28/go-darkpanda-backend/mock"
	"github.com/stretchr/testify/suite"
)

type EmitServiceConfirmMessageTestSuite struct {
	suite.Suite
	depCon container.Container
}

func (suite *EmitServiceConfirmMessageTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			suite.depCon = deps.Get().Container
		})
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

	iqParams, err := util.GenTestInquiryParams(maleUser.ID)

	if err != nil {
		suite.T().Fatal(err)
	}

	iq, err := q.CreateInquiry(ctx, *iqParams)

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

	mUserDao := mock.NewMockUserDAOer(ctrl)
	mServiceDao := mock.NewMockServiceDAOer(ctrl)
	mIqDao := mock.NewMockInquiryDAOer(ctrl)

	suite.depCon.Transient(func() contracts.UserDAOer {
		return mUserDao
	})

	suite.depCon.Transient(func() contracts.InquiryDAOer {
		return mIqDao
	})

	serviceUUID := uuid.New()

	// Mock request user.
	mUserDao.
		EXPECT().
		GetUserByUuid(gomock.Eq(maleUser.Uuid), gomock.Eq("username"), gomock.Eq("id")).
		Return(
			&maleUser,
			nil,
		)

	// Mock service.
	mServiceDao.EXPECT().
		GetServiceByInquiryUUID(gomock.Eq("some_inquiry_uuid")).
		Return(
			&models.Service{
				Uuid: sql.NullString{
					Valid:  true,
					String: serviceUUID.String(),
				},
			},
			nil,
		)

	// Mock inquiry.
	mIqDao.
		EXPECT().
		GetInquiryByUuid(gomock.Eq("some_inquiry_uuid")).
		Return(
			&contracts.InquiryResult{
				iq,
				maleUser.Username,
				maleUser.Uuid,
				maleUser.AvatarUrl,
			},
			nil,
		)

	c.Set("uuid", maleUser.Uuid)
	chat.EmitServiceConfirmedMessage(c, suite.depCon)
	apperr.HandleError()(c)

	log.Printf("RES %v", string(w.Body.Bytes()))

	// Retrieve message from firestore to makesure the correctness of the content.
	// resp := struct {
	// 	Message   interface{} `json:"message"`
	// 	MessageID string      `json:"message_id"`
	// }{}

	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
	// 	suite.T().Fatal(err)
	// }

	// _, err = darkfirestore.
	// 	Get().
	// 	Client.
	// 	Collection(darkfirestore.PrivateChatsCollectionName).
	// 	Doc("some_channel_uuid").
	// 	Collection(darkfirestore.MessageSubCollectionName).
	// 	Doc(resp.MessageID).
	// 	Get(ctx)

	// if err != nil {
	// 	suite.T().Fatal(err)
	// }

	// assert := assert.New(suite.T())
	// assert.Nil(err)

	// _, err = darkfirestore.
	// 	Get().
	// 	Client.
	// 	Collection(darkfirestore.PrivateChatsCollectionName).
	// 	Doc("some_channel_uuid").
	// 	Collection(darkfirestore.MessageSubCollectionName).
	// 	Doc(resp.MessageID).
	// 	Delete(ctx)

}

// This test should be skipped.
// func (suite *EmitServiceConfirmMessageTestSuite) TestEmitServiceConfirmedMessageSuccessToExistingService() {
// 	// Retrieve male user from real application

// 	// Make a real http request to existing server.
// 	headerMap := util.CreateJwtHeaderMap(
// 		"31a6b0dc-2857-4bad-b18e-76caab794dee",
// 		config.GetAppConf().JwtSecret,
// 	)

// 	v := &url.Values{}
// 	v.Add("price", "10.2")
// 	v.Add("channel_uuid", "private_chat:tq9MY5hGR")
// 	v.Add("inquiry_uuid", "d731a4f9-6907-4ca7-87ca-72b04df03ca8")
// 	v.Add("service_time", time.Now().Format("2006-01-02T15:04:05Z07:00"))
// 	v.Add("service_duration", "30")
// 	v.Add("service_type", string(models.ServiceTypeSex))

// 	req, err := http.NewRequest(
// 		"POST",
// 		"http://localhost:3001/v1/chat/emit-service-confirmed-message",
// 		strings.NewReader(v.Encode()),
// 	)

// 	if err != nil {
// 		suite.T().Fatal(err)
// 	}

// 	util.MergeFormUrlEncodedToHeader(req, headerMap)

// 	client := &http.Client{}
// 	res, err := client.Do(req)

// 	if err != nil {
// 		suite.T().Fatal(err)
// 	}

// 	resBytes, err := ioutil.ReadAll(res.Body)

// 	if err != nil {
// 		suite.T().Fatal(err)
// 	}

// 	log.Printf("DEBUG resp %v", string(resBytes))
// }

func TestEmitServiceConfirmMessageTestSuite(t *testing.T) {
	suite.Run(t, new(EmitServiceConfirmMessageTestSuite))
}
