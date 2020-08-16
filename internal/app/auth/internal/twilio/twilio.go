package twilio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Implementation of Twlio client. Please refer to this [documentation](https://www.twilio.com/docs/iam/test-credentials) to implement
// tests for various senarios sending SMS message. Twilio provides different 'from' mobile number that returns different response type when
// sending SMS message. Use those mobile numbers to implement tests handling various errors.

// To use this client in production, change the twilio credential to `LIVE Credentials` and use the mobile number states at your dashboard page.

const (
	TwilioBaseAPI = "https://api.twilio.com"

	// CreateMessage URL to create a message resource to send a new message
	// @ref: https://www.twilio.com/docs/sms/api/message-resource
	CreateMessage = "2010-04-01/Accounts/%s/Messages.json"
)

// SMSError error struct that will be returned when the status code of the response from twilio API is not within the range of 200 ~ 299
// Sample response format: {"code": 20003, "detail": "", "message": "Authenticate", "more_info": "https://www.twilio.com/docs/errors/20003", "status": 401
type SMSError struct {
	Code    int    `json:"code"`
	Detail  string `json:"detail"`
	Message string `json:"message"`
}

func NewSMSError(resReader io.Reader) *SMSError {
	dec := json.NewDecoder(resReader)

	var e SMSError
	if err := dec.Decode(&e); err != nil {
		log.WithFields(log.Fields{
			"message": "failed to decode twilio error response",
		}).Fatal(err)
	}

	return &e
}

func (e *SMSError) Error() string {
	return fmt.Sprintf("code: %d, detail %s, message %s", e.Code, e.Detail, e.Message)
}

type SMSResponse struct {
	SID string `json:"sid"`
}

func NewSMSResponse(resReader io.Reader) *SMSResponse {
	dec := json.NewDecoder(resReader)

	var r SMSResponse
	if err := dec.Decode(&r); err != nil {
		log.WithFields(log.Fields{
			"message": "failed to decode twilio sms response",
		}).Fatal(err)
	}

	return &r
}

type TwilioConf struct {
	AccountSID   string
	AccountToken string
}

type TwilioClient struct {
	Conf TwilioConf
}

func New(conf TwilioConf) *TwilioClient {
	return &TwilioClient{
		Conf: conf,
	}
}

func (tc *TwilioClient) getSendSMSUrl() string {
	u, _ := url.Parse(TwilioBaseAPI)
	u.Path = path.Join(u.Path, CreateMessage)

	return fmt.Sprintf(u.String(), tc.Conf.AccountSID)
}

func (tc *TwilioClient) SendSMS(from string, to string, content string) (*SMSResponse, error) {
	// Build out the data for our message
	v := url.Values{}
	v.Set("To", to)
	v.Set("From", from)
	v.Set("Body", content)
	rb := *strings.NewReader(v.Encode())

	// create client
	client := &http.Client{}
	req, _ := http.NewRequest("POST", tc.getSendSMSUrl(), &rb)
	req.SetBasicAuth(
		tc.Conf.AccountSID,
		tc.Conf.AccountToken,
	)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// make a request
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// if response status in withing 200 ~ 299 unmarshal body with success response struct.
	// otherwise, unmarshal it with error strcut.
	if resp.StatusCode >= http.StatusOK && resp.StatusCode <= 299 {
		return NewSMSResponse(resp.Body), nil
	}

	return nil, NewSMSError(resp.Body)
}

//func (tc *TwilioClientN)
