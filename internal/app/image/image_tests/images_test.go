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
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ImageTestSuite struct {
	suite.Suite

	sendRequest util.SendRequest
}

func (suite *ImageTestSuite) SetupSuite() {
	manager.NewDefaultManager()
	suite.sendRequest = util.SendRequestToApp(app.StartApp(gin.Default()))
}

func (suite *ImageTestSuite) TestUploadImage() {
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
	femaleUserParam, _ := util.GenTestUserParams(ctx)
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

func TestImageSuite(t *testing.T) {
	suite.Run(t, new(ImageTestSuite))
}
