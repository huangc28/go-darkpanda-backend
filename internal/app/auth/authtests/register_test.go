package authtests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"
)

type UserRegistrationTestSuite struct {
	suite.Suite
}

func (suite *UserRegistrationTestSuite) SetupSuite() {
	manager.NewDefaultManager()
}

func (suite *UserRegistrationTestSuite) TearDownSuite() {
	// @TODO close database connection
	log.Println("teardown test suite")
}

func (suite *UserRegistrationTestSuite) TestRegisterMissingParams() {
	suite.T().Skip()
	const ReferCode = "somerefercode"
	body := struct {
		ReferCode string `json:"refer_code"`
	}{
		ReferCode,
	}

	bodyB, _ := json.Marshal(&body)
	req, err := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(bodyB))

	if err != nil {
		suite.T().Fatalf("[register_missing_params] failed to request registerAPI %s", err.Error())
	}

	router := app.StartApp(gin.Default())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(suite.T(), rr.Result().StatusCode, http.StatusBadRequest)
	r := json.NewDecoder(rr.Body)
	resStruct := struct {
		ErrCode string `json:"err_code"`
	}{}

	r.Decode(&resStruct)
	assert.Equal(suite.T(), resStruct.ErrCode, apperr.FailedToValidateRegisterParams)
}

func (suite *UserRegistrationTestSuite) TestRegisterApiSuccess() {
	q := models.New(db.GetDB())

	// Create invitor
	usr, err := q.CreateUser(context.Background(), models.CreateUserParams{
		Username: "Bryan Huang",
		PhoneVerified: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		AuthSmsCode: sql.NullInt32{
			Int32: 3333,
			Valid: true,
		},
		Gender:      models.GenderFemale,
		PremiumType: models.PremiumTypeNormal,
		PremiumExpiryDate: sql.NullTime{
			Valid: false,
		},
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	// Create refer code data
	const ReferCode = "somerefercode"
	_, err = q.CreateRefcode(context.Background(), models.CreateRefcodeParams{
		InvitorID: int32(usr.ID),
		InviteeID: sql.NullInt32{
			Valid: false,
		},
		RefCode:     ReferCode,
		RefCodeType: models.RefCodeTypeInvitor,
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	body := struct {
		ReferCode string `json:"refer_code"`
		Username  string `json:"username"`
		Gender    string `json:"gender"`
	}{
		ReferCode,
		"bryan huang",
		"female",
	}

	bodyB, _ := json.Marshal(&body)

	req, err := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(bodyB))

	if err != nil {
		suite.T().Fatal(err)
	}

	router := app.StartApp(gin.Default())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	//log.Printf("rr status %d", rr.Code)
}

func TestRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserRegistrationTestSuite))
}
