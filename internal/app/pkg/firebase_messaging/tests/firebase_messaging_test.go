package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type FirebaseMessagingTestSuite struct {
	suite.Suite
	app *firebase.App
}

func (s *FirebaseMessagingTestSuite) SetupSuite() {
	ctx := context.Background()

	manager.NewDefaultManager(ctx).Run(func() {
		// cfile := fmt.Sprintf("%s/%s", config.GetProjRootPath(), "service_account_2.json")
		// cfile := fmt.Sprintf("%s/%s", config.GetProjRootPath(), "google-service.json")
		cfile := fmt.Sprintf("%s/%s", config.GetProjRootPath(), "dark-panda-gcp-service-account.json")

		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfile)
		app, err := firebase.NewApp(ctx, &firebase.Config{
			ProjectID: "dark-panda-6fe35",
		}, option.WithCredentialsFile(cfile))

		if err != nil {
			s.T().Fatal(err)
		}

		s.app = app
	})
}

func (s *FirebaseMessagingTestSuite) TestSendFirebaseMessaging() {
	ctx := context.Background()
	client, err := s.app.Messaging(ctx)

	if err != nil {
		s.T().Fatal(err)
	}

	m, err := client.Send(ctx, &messaging.Message{
		Topic: "inquiry_ZtNUKr77g_1629725654",
		Notification: &messaging.Notification{
			Title: "tester message",
			Body:  "hello yap",
		},
		Data: map[string]string{
			"title":   "tester message from data",
			"content": "hello yap from data",
		},
	})

	if err != nil {
		s.T().Fatal(err)
	}

	log.Printf("DEBUG m %v", m)
}

func TestFirebaseMessagingSuite(t *testing.T) {
	suite.Run(t, new(FirebaseMessagingTestSuite))
}
