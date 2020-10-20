package inquirytests

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type PickupInquiryTestSuite struct {
	suite.Suite
	SendUrlEncodedRequest util.SendUrlEncodedRequest
}

func (suite *PickupInquiryTestSuite) SetupSuite() {
	manager.NewDefaultManager()
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
	iqParams.InquiryStatus = models.InquiryStatusInquiring
	iqParams.ServiceType = models.ServiceTypeSex
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

	//log.Printf("DEBUG female user %v", femaleUser)

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

	log.Printf("DEBUG 3 %s", string(resp.Body.Bytes()))

	// ------------------- Assert test cases -------------------
	//assert.Equal(suite.T(), http.StatusOK, resp.Result().StatusCode)

	//respBody := inquiry.TransformedPickupInquiry{}
	//dec := json.NewDecoder(resp.Body)
	//if err := dec.Decode(&respBody); err != nil {
	//suite.T().Fatal(err)
	//}

	//assert.NotEmpty(suite.T(), respBody.Uuid)
	//assert.Equal(suite.T(), string(models.ServiceTypeSex), respBody.ServiceType)
	//assert.Equal(suite.T(), string(models.InquiryStatusChatting), respBody.InquiryStatus)

	//assert.NotEmpty(suite.T(), respBody.Inquirer.Uuid)
	//assert.Equal(suite.T(), maleUser.Username, respBody.Inquirer.Username)
	//assert.Equal(suite.T(), string(maleUser.PremiumType), respBody.Inquirer.PremiumType)
}

func TestPickupInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(PickupInquiryTestSuite))
}
