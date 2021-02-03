package tests

import (
	"context"
	"testing"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/twilio"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// initialize application config
// use config to initialize twilio client
type TwilioTestSuite struct {
	suite.Suite
	twilioConf   config.TwilioConf
	twilioClient *twilio.TwilioClient
}

func (suite *TwilioTestSuite) SetupSuite() {
	manager.NewDefaultManager(context.Background())

	twilioConf := config.GetAppConf().TwilioConf
	suite.twilioClient = twilio.New(twilio.TwilioConf{
		AccountSID:   twilioConf.AccountId,
		AccountToken: twilioConf.AuthToken,
	})
}

func (suite *TwilioTestSuite) TestSendSMSSuccess() {
	resp, err := suite.twilioClient.SendSMS(
		"+15005550006",
		"+886988272727",
		"test from darkpanda",
	)

	if err != nil {
		suite.T().Fatal(err)
	}

	assert.NotEmpty(suite.T(), resp.SID)
}

func (suite *TwilioTestSuite) TestSendSMSFailedDueToInvalidPhoneNumber() {
	_, err := suite.twilioClient.SendSMS(
		"+15005550001",
		"+886988272727",
		"test from darkpanda",
	)

	SMSErr := err.(*twilio.SMSError)
	assert.Equal(suite.T(), SMSErr.Code, 21212)
}

func TestTwilioTestSuite(t *testing.T) {
	suite.Run(t, new(TwilioTestSuite))
}
