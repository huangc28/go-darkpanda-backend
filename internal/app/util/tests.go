package util

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	faker "github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/ventu-io/go-shortid"
)

func randomGender() models.Gender {
	gs := []models.Gender{
		models.GenderFemale,
		models.GenderMale,
	}

	return gs[seededRand.Intn(len(gs))]
}

// GenTestUser generate randomized data on user fields but now create it.
func GenTestUserParams(ctx context.Context) (*models.CreateUserParams, error) {
	p := &models.CreateUserParams{}
	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	sid, err := shortid.Generate()

	if err != nil {
		return nil, err
	}

	p.Username = faker.Username()
	p.Uuid = sid
	p.Gender = randomGender()
	p.PremiumType = models.PremiumTypeNormal
	p.PhoneVerifyCode = sql.NullString{
		String: fmt.Sprintf("%s-%d", GenRandStringRune(3), Gen4DigitNum(1000, 9999)),
		Valid:  true,
	}

	return p, nil
}

func randomServiceType() models.ServiceType {
	gs := []models.ServiceType{
		models.ServiceTypeChat,
		models.ServiceTypeDiner,
		models.ServiceTypeMovie,
		models.ServiceTypeSex,
		models.ServiceTypeShopping,
	}

	return gs[seededRand.Intn(len(gs))]
}

func randomInquiryStatus() models.InquiryStatus {
	gs := []models.InquiryStatus{
		models.InquiryStatusBooked,
		models.InquiryStatusCanceled,
		models.InquiryStatusExpired,
		models.InquiryStatusInquiring,
	}

	return gs[seededRand.Intn(len(gs))]
}

func GenTestInquiryParams(inquirerID int64) (*models.CreateInquiryParams, error) {
	p := &models.CreateInquiryParams{}
	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	p.InquirerID = sql.NullInt32{
		Int32: int32(inquirerID),
		Valid: true,
	}
	p.ServiceType = randomServiceType()
	p.InquiryStatus = randomInquiryStatus()

	return p, nil
}

type SendRequest func(method string, url string, body interface{}, header map[string]string) (*httptest.ResponseRecorder, error)

func SendRequestToApp(e *gin.Engine) SendRequest {
	return func(method string, url string, body interface{}, header map[string]string) (*httptest.ResponseRecorder, error) {

		bbody, err := json.Marshal(&body)

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest(method, url, bytes.NewBuffer(bbody))

		for headerKey, headerVal := range header {
			req.Header.Set(headerKey, headerVal)
		}

		if err != nil {
			return nil, err
		}

		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		return rr, nil
	}
}

func CreateJwtHeaderMap(uuid, secret string) map[string]string {
	header := make(map[string]string)
	token, _ := jwtactor.CreateToken(uuid, secret)
	header["Authorization"] = fmt.Sprintf("Bearer %s", token)

	return header
}
