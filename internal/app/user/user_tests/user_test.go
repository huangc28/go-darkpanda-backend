package usertests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/image"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserAPITestsSuite struct {
	suite.Suite
	sendRequest           util.SendRequest
	sendUrlEncodedRequest util.SendUrlEncodedRequest
}

func (suite *UserAPITestsSuite) SetupSuite() {
	manager.
		NewDefaultManager().
		Run(func() {
			tApp := app.StartApp(gin.Default())
			suite.sendRequest = util.SendRequestToApp(tApp)
			suite.sendUrlEncodedRequest = util.SendUrlEncodedRequestToApp(tApp)
		})

}

func (suite *UserAPITestsSuite) TestGetMaleUserInfo() {
	// create a male user that has no related active inquiry
	ctx := context.Background()
	newUserParams, err := util.GenTestUserParams()
	newUserParams.Gender = models.GenderMale

	if err != nil {
		suite.T().Fatal(err)
	}

	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- Create jwt to request the API -------------------
	jwt, err := jwtactor.CreateToken(newUser.Uuid, config.GetAppConf().JwtSecret)

	if err != nil {
		suite.T().Fatal(err)
	}

	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", jwt)

	resp, _ := suite.sendRequest("GET", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases  -------------------
	assert.Equal(suite.T(), resp.Result().StatusCode, http.StatusOK)
	dec := json.NewDecoder(resp.Result().Body)
	respStruct := &user.TransformUserWithInquiryData{}
	if err := dec.Decode(&respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), respStruct.Username, newUser.Username)
	assert.Equal(suite.T(), respStruct.Gender, newUser.Gender)
	assert.Equal(suite.T(), respStruct.Uuid, newUser.Uuid)

	assert.Equal(suite.T(), len(respStruct.Inquiries), 0)

}

func (suite *UserAPITestsSuite) TestGetMaleUserInfoWithActiveInquiry() {
	// create a male user along with active inquiries
	ctx := context.Background()
	newUserParams, err := util.GenTestUserParams()

	if err != nil {
		suite.T().Fatal(err)
	}

	newUserParams.Gender = models.GenderMale

	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *newUserParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	newInquiryParams, err := util.GenTestInquiryParams(newUser.ID)
	if err != nil {
		suite.T().Fatal(err)
	}
	newInquiryParams.InquiryStatus = models.InquiryStatusInquiring
	if _, err := q.CreateInquiry(ctx, *newInquiryParams); err != nil {
		suite.T().Fatalf("Failed to create inquiry %s", err.Error())
	}

	// ------------------- Create jwt to request the API -------------------
	jwt, err := jwtactor.CreateToken(newUser.Uuid, config.GetAppConf().JwtSecret)

	if err != nil {
		suite.T().Fatal(err)
	}

	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", jwt)
	resp, _ := suite.sendRequest("GET", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases -------------------
	respStruct := &user.TransformUserWithInquiryData{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(suite.T(), len(respStruct.Inquiries), 1)
}

func (suite *UserAPITestsSuite) TestGetFemaleUserInfo() {
	ctx := context.Background()
	userParams, _ := util.GenTestUserParams()
	userParams.Gender = models.GenderFemale

	q := models.New(db.GetDB())
	newUser, err := q.CreateUser(ctx, *userParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// request the api

	header := util.CreateJwtHeaderMap(
		newUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, _ := suite.sendRequest("GET", "/v1/users/me", struct{}{}, header)

	// ------------------- assert test cases -------------------
	respStruct := &user.TransformedUser{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(respStruct.Uuid, newUser.Uuid)
	assert.Equal(respStruct.AvatarUrl, newUser.AvatarUrl.String)
	assert.Equal(respStruct.Username, newUser.Username)
	assert.Equal(respStruct.Gender, newUser.Gender)
}

func (suite *UserAPITestsSuite) TestPutUserInfoSuccess() {
	// ------------------- create male user -------------------
	ctx := context.Background()
	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = models.GenderMale
	q := models.New(db.GetDB())
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- send API -------------------
	headers := util.CreateJwtHeaderMap(
		maleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	body := struct {
		AvatarURL   string  `json:"avatar_url"`
		Nationality string  `json:"nationality"`
		Region      string  `json:"region"`
		Height      float32 `json:"height"`
		Weight      float32 `json:"weight"`
		Age         int     `json:"age"`
		Description string  `json:"description"`
		BreastSize  string  `json:"breast_size"`
	}{
		AvatarURL:   "https://somecloud.com/a.png",
		Nationality: "Taiwan",
		Region:      "Taipei",
		Height:      170,
		Weight:      68,
		Age:         18,
		Description: "I am rich and have big dick",
		BreastSize:  "z",
	}

	res, err := suite.sendRequest(
		"PUT",
		fmt.Sprintf("/v1/users/%s", maleUser.Uuid),
		body,
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	if res.Result().StatusCode != http.StatusOK {
		bbody, _ := ioutil.ReadAll(res.Result().Body)
		suite.T().Fatalf("%s", string(bbody))
	}

	// ------------------- test asserts -------------------
	assert := assert.New(suite.T())
	dec := json.NewDecoder(res.Result().Body)
	resBody := struct {
		AvatarURL   *string `json:"avatar_url"`
		Nationality *string `json:"nationality"`
		Age         *int    `json:"age"`
	}{}

	if err := dec.Decode(&resBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(*resBody.AvatarURL, body.AvatarURL)
	assert.Equal(*resBody.Nationality, "Taiwan")
	assert.Equal(*resBody.Age, 18)

	// ------------------- test nil update -------------------
	resTwo, err := suite.sendRequest(
		"PUT",
		fmt.Sprintf("/v1/users/%s", maleUser.Uuid),
		struct {
			Age int `json:"age"`
		}{
			19,
		},
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	resTwoBody := struct {
		AvatarURL   *string `json:"avatar_url"`
		Nationality *string `json:"nationality"`
		Age         *int    `json:"age"`
	}{}
	dec = json.NewDecoder(resTwo.Result().Body)
	if err := dec.Decode(&resTwoBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(*resBody.AvatarURL, body.AvatarURL)
	assert.Equal(*resBody.Nationality, "Taiwan")
	assert.Equal(*resTwoBody.Age, 19)
}

func (suite *UserAPITestsSuite) TestGetUserProfileByUuid() {
	// create a female user
	ctx := context.Background()
	q := models.New(db.GetDB())

	femaleUserParams, _ := util.GenTestUserParams()
	femaleUserParams.Gender = "female"
	femaleUserParams.PhoneVerified = true
	femaleUserParams.Mobile = sql.NullString{
		Valid:  true,
		String: "+886988272727",
	}

	femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

	if err != nil {
		suite.T().Fatalf("create female user %s", err.Error())
	}

	// create a male user
	maleUserParams, _ := util.GenTestUserParams()
	maleUserParams.Gender = "male"
	maleUser, err := q.CreateUser(ctx, *maleUserParams)
	maleUser.PhoneVerified = true
	maleUser.Mobile = sql.NullString{
		Valid:  true,
		String: "+886986900050",
	}

	if err != nil {
		suite.T().Fatalf("create male user %s", err.Error())
	}

	// female user wants to view the profile of a male user
	headers := util.CreateJwtHeaderMap(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, err := suite.sendUrlEncodedRequest(
		"GET",
		fmt.Sprintf("/v1/users/%s", maleUser.Uuid),
		&url.Values{},
		headers,
	)

	if resp.Code != http.StatusOK {
		suite.T().Fatal(err)
	}

	// ------------------- test cases -------------------
	respStruct := &user.TransformedMaleUser{}
	dec := json.NewDecoder(resp.Result().Body)
	if err := dec.Decode(respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(maleUser.Uuid, respStruct.Uuid)
	assert.Equal(maleUser.Username, respStruct.Username)
}

func (suite *UserAPITestsSuite) TestGetUserImagesByUuid() {
	// create a male user
	maleUserParams, err := util.GenTestUserParams()
	ctx := context.Background()

	if err != nil {
		log.Fatal(err)
	}

	maleUserParams.Gender = "male"
	q := models.New(db.GetDB())
	maleUser, err := q.CreateUser(ctx, *maleUserParams)

	if err != nil {
		log.Fatal(err)
	}

	imgDao := image.ImageDAO{
		DB: db.GetDB(),
	}

	// create images and relate those images to that user
	imagesParams := make([]image.CreateImageParams, 0)
	for i := 0; i < 12; i++ {
		imagesParams = append(imagesParams, image.CreateImageParams{
			UserID: maleUser.ID,
			URL:    fmt.Sprintf("https://foo.com/bar%d.png", i),
		})
	}

	if err := imgDao.CreateImages(imagesParams); err != nil {
		log.Fatalf("Failed to insert images %s", err.Error())
	}

	// ------------------- request API -------------------
	headers := util.CreateJwtHeaderMap(maleUser.Uuid, config.GetAppConf().JwtSecret)
	resp, err := suite.sendUrlEncodedRequest(
		"GET",
		fmt.Sprintf("/v1/users/%s/images", maleUser.Uuid),
		&url.Values{},
		headers,
	)

	if err != nil {
		log.Fatal(err)
	}

	imgsStruct := user.TransformedUserImages{}
	if err := json.Unmarshal(resp.Body.Bytes(), &imgsStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, resp.Result().StatusCode)

	// 9 images per request
	assert.Equal(9, len(imgsStruct.Images))
}

func (suite *UserAPITestsSuite) TestGetUserPaymentSuccess() {
	// seed male / female user
	maleUserParams, err := util.GenTestUserParams()

	if err != nil {
		log.Fatal(err)
	}

	maleUserParams.Gender = "male"

	femaleUserParams, err := util.GenTestUserParams()

	if err != nil {
		log.Fatal(err)
	}

	femaleUserParams.Gender = "female"
	ctx := context.Background()
	q := models.New(db.GetDB())

	maleUser, _ := q.CreateUser(ctx, *maleUserParams)
	femaleUser, _ := q.CreateUser(ctx, *femaleUserParams)

	// seed an inquiry
	inquiryParams, err := util.GenTestInquiryParams(maleUser.ID)
	inquiry, err := q.CreateInquiry(ctx, *inquiryParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// seed a service that the female user has completed a service with the male user
	serviceParams, _ := util.GenTestServiceParams(maleUser.ID, femaleUser.ID, inquiry.ID)
	serviceParams.ServiceStatus = models.ServiceStatusCompleted
	service, err := q.CreateService(ctx, *serviceParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// seed a payment
	paymentParams, err := util.GenTestPayment(
		maleUser.ID,
		femaleUser.ID,
		service.ID,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	newPayment, err := q.CreatePayment(ctx, *paymentParams)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- sends API -------------------
	headers := util.CreateJwtHeaderMap(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, err := suite.sendUrlEncodedRequest(
		"GET",
		fmt.Sprintf("/v1/users/%s/payments", maleUser.Uuid),
		&url.Values{},
		headers,
	)

	if err != nil {
		log.Fatal(err)
	}

	// ------------------- assert test cases -------------------
	respStruct := user.TransformedPaymentInfos{}
	if err := json.Unmarshal(resp.Body.Bytes(), &respStruct); err != nil {
		log.Fatal(err)
	}
	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, resp.Result().StatusCode)
	assert.Equal(newPayment.RecTradeID.String, respStruct.Payments[0].RecTradeID)
}

func (suite *UserAPITestsSuite) TestGetUserHistoricalServices() {
	// Seed one male users as the customer with multiple completed services
	ctx := context.Background()
	q := models.New(db.GetDB())

	maleUserParams, err := util.GenTestUserParams()

	if err != nil {
		log.Fatal(err)
	}

	maleUserParams.Gender = "male"
	maleUser, err := q.CreateUser(ctx, *maleUserParams)

	if err != nil {
		log.Fatal(err)
	}

	// seed multiple female users that had hooked up with the male user
	femaleUsers := make([]models.User, 0)
	for i := 0; i < 6; i++ {
		femaleUserParams, err := util.GenTestUserParams()

		if err != nil {
			log.Fatal(err)
		}

		femaleUserParams.Gender = "female"
		femaleUser, err := q.CreateUser(ctx, *femaleUserParams)

		if err != nil {
			log.Fatal(err)
		}

		femaleUsers = append(femaleUsers, femaleUser)
	}

	// create inquiries for the male user
	inquiries := make([]models.ServiceInquiry, 0)
	for range femaleUsers {
		iqParams, err := util.GenTestInquiryParams(maleUser.ID)

		if err != nil {
			log.Fatal(err)
		}

		iq, err := q.CreateInquiry(ctx, *iqParams)

		if err != nil {
			log.Fatal(err)
		}

		inquiries = append(inquiries, iq)
	}

	// create services for the male user
	for idx, inquiry := range inquiries {
		serviceParams, err := util.GenTestServiceParams(
			maleUser.ID,
			femaleUsers[idx].ID,
			inquiry.ID,
		)

		if err != nil {
			log.Fatal(err)
		}

		serviceParams.ServiceStatus = models.ServiceStatusCompleted
		_, err = q.CreateService(ctx, *serviceParams)

		if err != nil {
			log.Fatal(err)
		}
	}

	// ------------------- request API -------------------
	headers := util.CreateJwtHeaderMap(
		femaleUsers[0].Uuid,
		config.GetAppConf().JwtSecret,
	)

	resp, err := suite.sendUrlEncodedRequest(
		"GET",
		fmt.Sprintf("/v1/users/%s/services", maleUser.Uuid),
		&url.Values{},
		headers,
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- assertions -------------------
	respStruct := user.TransformedHistoricalServices{}
	if err := json.Unmarshal(resp.Body.Bytes(), &respStruct); err != nil {
		suite.T().Fatal(err)
	}

	assert := assert.New(suite.T())
	assert.Equal(http.StatusOK, resp.Result().StatusCode)
	assert.Equal(5, len(respStruct.Services), "5 records per page")
}

func TestUserAPISuite(t *testing.T) {
	suite.Run(t, new(UserAPITestsSuite))
}
