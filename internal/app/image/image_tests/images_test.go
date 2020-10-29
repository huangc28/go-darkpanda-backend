package imagetests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ImageTestSuite struct {
	suite.Suite
}

func (suite *ImageTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())
}

// Refer to this [article](https://stackoverflow.com/questions/26063271/how-to-create-a-http-request-that-contains-multiple-fileheaders)
// to test multiple file upload in a request.
func (suite *ImageTestSuite) TestUploadAvatar() {
	file, _ := os.Open("./download.png")
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "download.png")

	if _, err := io.Copy(part, file); err != nil {
		suite.T().Fatal(err)
	}

	writer.Close()

	ctx := context.Background()
	femaleUserParam, _ := util.GenTestUserParams()
	femaleUserParam.Gender = models.GenderFemale
	q := models.New(db.GetDB())
	femaleUser, err := q.CreateUser(ctx, *femaleUserParam)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- send API -------------------
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		"/v1/images/avatar",
		body,
	)

	jwtToken, _ := jwtactor.CreateToken(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	app.StartApp(gin.Default()).ServeHTTP(res, req)

	// ------------------- assert test cases -------------------
	assert := assert.New(suite.T())
	assert.Equal(res.Result().StatusCode, http.StatusOK)

	respBody := struct {
		PublicLink string `json:"public_link"`
	}{}

	dec := json.NewDecoder(res.Result().Body)
	if err := dec.Decode(&respBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal("https://storage.googleapis.com/petu-love.appspot.com/download.png", respBody.PublicLink)
}

func (suite *ImageTestSuite) TestUploadMultipleImages() {
	// ------------------- create multiparts -------------------
	file1, _ := os.Open("./download.png")
	defer file1.Close()

	file2, _ := os.Open("./download3.png")
	defer file2.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	p1, _ := writer.CreateFormFile("image", file1.Name())

	if _, err := io.Copy(p1, file1); err != nil {
		suite.T().Fatalf("failed to copy file1 to part 1 %s", err.Error())
	}

	p2, _ := writer.CreateFormFile("image", file2.Name())
	if _, err := io.Copy(p2, file2); err != nil {
		suite.T().Fatalf("failed to copy file2 to part 2 %s", err.Error())
	}

	writer.Close()

	// ------------------- create test users -------------------
	ctx := context.Background()
	femaleUserParam, _ := util.GenTestUserParams()
	femaleUserParam.Gender = models.GenderFemale
	q := models.New(db.GetDB())
	femaleUser, err := q.CreateUser(ctx, *femaleUserParam)

	if err != nil {
		suite.T().Fatal(err)
	}

	// ------------------- send API -------------------
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		"/v1/images",
		body,
	)

	jwtToken, _ := jwtactor.CreateToken(
		femaleUser.Uuid,
		config.GetAppConf().JwtSecret,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	app.StartApp(gin.Default()).ServeHTTP(res, req)

	respBody := struct {
		Links []string `json:"links"`
	}{}

	assert := assert.New(suite.T())

	assert.Equal(http.StatusOK, res.Result().StatusCode)
	dec := json.NewDecoder(res.Result().Body)
	if err := dec.Decode(&respBody); err != nil {
		suite.T().Fatal(err)
	}

	assert.Equal(2, len(respBody.Links))
}

func TestImageSuite(t *testing.T) {
	suite.Run(t, new(ImageTestSuite))
}
