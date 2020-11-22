package chattests

import (
	"context"
	"database/sql"
	"log"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type GetInquiryChatTestSuite struct {
	suite.Suite
}

func (suite *GetInquiryChatTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
}

func (suite *GetInquiryChatTestSuite) TestGetInquiryChatSuccess() {

	q := models.New(db.GetDB())
	ctx := context.Background()
	// Seed several inquirers.
	inquirers := make([]models.User, 0)
	for i := 0; i < 3; i++ {
		maleUserParams, err := util.GenTestUserParams()
		maleUserParams.Gender = models.GenderMale
		maleUser, err := q.CreateUser(ctx, *maleUserParams)
		if err != nil {
			suite.T().Fatal(err)
		}

		inquirers = append(inquirers, maleUser)
	}

	// Seed several female users to pick inquiries
	femaleUserParams, err := util.GenTestUserParams()
	femaleUserParams.Gender = models.GenderFemale
	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	// Seed several inquiries
	inquiries := make([]models.ServiceInquiry, 0)
	for _, inquirer := range inquirers {
		inquiryParam, err := util.GenTestInquiryParams(inquirer.ID)
		inquiryParam.PickerID = sql.NullInt32{
			Valid: true,
			Int32: int32(femaleUser.ID),
		}

		inquiryParam.ServiceType = models.ServiceTypeSex
		inquiryParam.InquiryStatus = models.InquiryStatusChatting

		if err != nil {
			suite.T().Fatal(err)
		}

		inquiry, err := q.CreateInquiry(ctx, *inquiryParam)
		if err != nil {
			suite.T().Fatal(err)
		}
		inquiries = append(inquiries, inquiry)
	}

	// Seed Chatrooms
	for _, inquiry := range inquiries {
		chatroomParams, err := util.GenTestChat(
			inquiry.ID,
			[]int64{
				femaleUser.ID,
				int64(inquiry.InquirerID.Int32),
			}...,
		)

		if err != nil {
			suite.T().Fatal(err)
		}

		if _, err := q.CreateChatroom(ctx, *chatroomParams); err != nil {
			if err != nil {
				suite.T().Fatal(err)
			}
		}
	}

	// Request API
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uuid", femaleUser.Uuid)

	req, err := util.ComposeTestRequest(
		"GET",
		"/inquiry-chatrooms",
		&url.Values{},
		util.CreateJwtHeaderMap(
			femaleUser.Uuid,
			config.GetAppConf().JwtSecret,
		),
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	handlers := chat.ChatHandlers{
		UserDao: user.NewUserDAO(db.GetDB()),
		ChatDao: chat.NewChatDao(db.GetDB()),
	}

	c.Request = req
	handlers.GetInquiryChatRooms(c)
	apperr.HandleError()(c)

	// assertions
	// respStruct := models.InquiryChatRooms{}
	// if err := json.Unmarshal(w.Body.Bytes(), &respStruct); err != nil {
	// 	if err != nil {
	// 		suite.T().Fatal(err)
	// 	}
	// }

	log.Printf("DEBUG  2 %v", string(w.Body.Bytes()))
}

func TestGetInquiryChatTestSuite(t *testing.T) {
	suite.Run(t, new(GetInquiryChatTestSuite))
}
