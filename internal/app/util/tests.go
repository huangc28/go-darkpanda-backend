package util

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	faker "github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/ventu-io/go-shortid"
)

func randomBool() bool {
	return seededRand.Intn(2) == 1
}

func randomGender() models.Gender {
	gs := []models.Gender{
		models.GenderFemale,
		models.GenderMale,
	}

	return gs[seededRand.Intn(len(gs))]
}

// GenTestUser generate randomized data on user fields but now create it.
// @TODO
//   - remove unecessary argument `ctx`
func GenTestUserParams() (*models.CreateUserParams, error) {
	p := &models.CreateUserParams{}
	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	sid, err := shortid.Generate()

	if err != nil {
		return nil, err
	}

	p.Username = faker.Username()
	p.Description = sql.NullString{
		Valid:  true,
		String: faker.Sentence(),
	}
	p.Uuid = sid

	p.Gender = randomGender()
	p.PhoneVerified = randomBool()
	p.Mobile = sql.NullString{
		Valid:  randomBool(),
		String: faker.Phonenumber(),
	}
	p.PremiumType = models.PremiumTypeNormal
	p.AvatarUrl = sql.NullString{
		Valid:  true,
		String: faker.URL(),
	}
	p.FcmTopic = sql.NullString{
		Valid:  true,
		String: "sampleFcmTopic",
	}

	return p, nil
}

type TestServiceInfo struct {
	Male    *models.User
	Female  *models.User
	Inquiry *models.ServiceInquiry
	Service *models.Service
}

type MalePreCreateHook func(*models.CreateUserParams)
type FemalePreCreateHook func(*models.CreateUserParams)
type InquiryPreCreateHook func(*models.CreateInquiryParams, *models.User)
type ServicePreCreateHook func(*models.CreateServiceParams)
type CreateTestServiceHooks struct {
	MalePreCreateHook    MalePreCreateHook
	FemalePreCreateHook  FemalePreCreateHook
	InquiryPreCreateHook InquiryPreCreateHook
	ServicePreCreateHook ServicePreCreateHook
}

// GenTestService generates test service with related inquiry with it's customer and service provider.
func CreateTestService(ctx context.Context, q *models.Queries, hooks CreateTestServiceHooks) (*TestServiceInfo, error) {
	maleParams, err := GenTestUserParams()

	if err != nil {
		return nil, err
	}

	maleParams.Gender = models.GenderMale

	if hooks.MalePreCreateHook != nil {
		hooks.MalePreCreateHook(maleParams)
	}

	male, err := q.CreateUser(ctx, *maleParams)

	if err != nil {
		return nil, err
	}

	femaleParams, err := GenTestUserParams()

	if err != nil {
		return nil, err
	}

	if hooks.FemalePreCreateHook != nil {
		hooks.FemalePreCreateHook(femaleParams)
	}

	female, err := q.CreateUser(ctx, *femaleParams)

	if err != nil {
		return nil, err
	}

	iqParams, err := GenTestInquiryParams(male.ID)

	if err != nil {
		return nil, err
	}

	if hooks.InquiryPreCreateHook != nil {
		hooks.InquiryPreCreateHook(iqParams, &female)
	}

	iq, err := q.CreateInquiry(ctx, *iqParams)

	if err != nil {
		return nil, err
	}

	serviceParams, err := GenTestServiceParams(male.ID, female.ID, iq.ID)

	if err != nil {
		return nil, err
	}

	if hooks.ServicePreCreateHook != nil {
		hooks.ServicePreCreateHook(serviceParams)
	}

	srv, err := q.CreateService(ctx, *serviceParams)

	if err != nil {
		return nil, err
	}

	return &TestServiceInfo{
		Male:    &male,
		Female:  &female,
		Inquiry: &iq,
		Service: &srv,
	}, nil
}

func randomSericeStatus() models.ServiceStatus {
	gs := []models.ServiceStatus{
		models.ServiceStatusUnpaid,
		models.ServiceStatusToBeFulfilled,
		models.ServiceStatusCanceled,
		models.ServiceStatusFulfilling,
		models.ServiceStatusExpired,
		models.ServiceStatusCompleted,
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

func randomFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + seededRand.Float64()*(max-min)
	}
	return res
}

func GenTestInquiryParams(inquirerID int64) (*models.CreateInquiryParams, error) {
	p := &models.CreateInquiryParams{}
	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	sid, _ := shortid.Generate()

	p.Uuid = sid
	p.InquirerID = sql.NullInt32{
		Int32: int32(inquirerID),
		Valid: true,
	}
	p.ExpectServiceType = sql.NullString{
		Valid:  true,
		String: faker.Word(),
	}
	p.InquiryStatus = randomInquiryStatus()
	p.Budget = fmt.Sprintf("%.2f", randomFloats(1.00, 102.99, 1)[0])
	p.Lat = sql.NullString{
		Valid: false,
	}
	p.Lng = sql.NullString{
		Valid: false,
	}

	p.AppointmentTime = sql.NullTime{
		Valid: true,
		Time:  time.Now().AddDate(0, 0, 1),
	}

	return p, nil
}

func GenTestServiceParams(customerID int64, serviceProviderID int64, inquiryID int64) (*models.CreateServiceParams, error) {
	p := &models.CreateServiceParams{}

	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	sid, _ := shortid.Generate()

	p.Uuid = sql.NullString{
		Valid:  true,
		String: sid,
	}

	p.CustomerID = sql.NullInt32{
		Int32: int32(customerID),
		Valid: true,
	}

	p.ServiceProviderID = sql.NullInt32{
		Int32: int32(serviceProviderID),
		Valid: true,
	}

	p.InquiryID = int32(inquiryID)

	p.Price = sql.NullString{
		String: fmt.Sprintf("%.2f", randomFloats(1.00, 102.99, 1)[0]),
		Valid:  true,
	}

	p.Budget = sql.NullString{
		String: fmt.Sprintf("%.2f", randomFloats(1.00, 102.99, 1)[0]),
		Valid:  true,
	}

	p.ServiceStatus = randomSericeStatus()

	p.ServiceType = faker.Word()

	return p, nil
}

func GenTestChat(inquiryID int64, chatUserIDs ...int64) (*models.CreateChatroomParams, error) {
	m := &models.CreateChatroomParams{}
	if err := faker.FakeData(m); err != nil {
		return nil, err
	}

	m.InquiryID = int32(inquiryID)
	m.ChannelUuid = sql.NullString{
		String: shortid.MustGenerate(),
		Valid:  true,
	}

	return m, nil
}

func GenTestPayment(payerID int64, serviceID int64) (*models.CreatePaymentParams, error) {
	p := &models.CreatePaymentParams{}

	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	p.PayerID = int32(payerID)
	p.ServiceID = int32(serviceID)

	p.Price = fmt.Sprintf("%.2f", randomFloats(1.00, 102.99, 1)[0])

	return p, nil
}

type SendRequest func(method string, url string, body interface{}, header map[string]string) (*httptest.ResponseRecorder, error)

func SendRequestToApp(e *gin.Engine) SendRequest {
	return func(method string, url string, body interface{}, header map[string]string) (*httptest.ResponseRecorder, error) {
		bbody, err := json.Marshal(&body)

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest(
			method,
			url,
			bytes.NewBuffer(bbody),
		)

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

type SendUrlEncodedRequest func(method string, url string, params *url.Values, headers map[string]string) (*httptest.ResponseRecorder, error)

func SendUrlEncodedRequestToApp(e *gin.Engine) SendUrlEncodedRequest {
	return func(method string, url string, params *url.Values, headers map[string]string) (*httptest.ResponseRecorder, error) {
		req, err := http.NewRequest(
			method,
			url,
			strings.NewReader(params.Encode()),
		)

		if err != nil {
			return nil, err
		}

		MergeFormUrlEncodedToHeader(req, headers)

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

func MergeMapToHeader(req *http.Request, headers map[string]string, contentType string) {
	for headerKey, headerVal := range headers {
		req.Header.Set(headerKey, headerVal)
	}

	req.Header.Set("Content-Type", contentType)
}

func MergeFormUrlEncodedToHeader(req *http.Request, headers map[string]string) {
	MergeMapToHeader(req, headers, "application/x-www-form-urlencoded")
}

func MergeJsonToHeader(req *http.Request, headers map[string]string) {
	MergeMapToHeader(req, headers, "application/json")
}

func ComposeTestRequest(method, url string, params *url.Values, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		url,
		strings.NewReader(params.Encode()),
	)

	if err != nil {
		return nil, err
	}

	MergeFormUrlEncodedToHeader(req, headers)

	return req, nil
}

func ComposeJsonTestRequest(method, url string, jsonBody interface{}, headers map[string]string) (*http.Request, error) {
	bbody, err := json.Marshal(&jsonBody)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(bbody),
	)

	if err != nil {
		return nil, err
	}

	MergeJsonToHeader(req, headers)

	return req, nil
}
