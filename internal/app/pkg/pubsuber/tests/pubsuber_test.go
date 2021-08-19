package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/pubsuber"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type PubsuberTestSuite struct {
	suite.Suite
	client *pubsub.Client
}

func (suite *PubsuberTestSuite) SetupSuite() {
	ctx := context.Background()
	manager.NewDefaultManager(ctx).Run(func() {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", fmt.Sprintf("%s/%s", config.GetProjRootPath(), "dark-panda-gcp-service-account.json"))

		client, err := pubsub.NewClient(ctx, config.GetAppConf().GcpProjectID)
		if err != nil {
			suite.T().Fatal(err)
		}

		suite.client = client
	})
}

func (suite *PubsuberTestSuite) TestCreateNewTopic() {
	ctx := context.Background()
	t, err := suite.client.CreateTopic(ctx, "example_topic")

	if err != nil {
		suite.T().Fatalf("failed to create topic %v", err)
	}

	tExists, err := t.Exists(ctx)

	if err != nil {
		suite.T().Fatalf("topic does not exist %v", err)
	}

	suite.Assertions.True(tExists)

	if err := t.Delete(ctx); err != nil {
		suite.T().Fatalf("failed to delete, please remove topic manually  %v", err)
	}
}

func (suite *PubsuberTestSuite) TestSubscribeToNewTopic() {
	// Create a new topic.
	ctx := context.Background()
	ps := pubsuber.New(suite.client)
	topic, err := ps.CreateInquiryTopic(ctx, "someinquiry")

	if err != nil {
		suite.T().Fatal(err)
	}

	sub, err := suite.client.CreateSubscription(ctx, fmt.Sprintf("%s_sub", topic.ID()), pubsub.SubscriptionConfig{
		Topic: topic,
	})

	if err != nil {
		suite.T().Fatal(err)
	}

	errChan := make(chan error)
	doneChan := make(chan []byte)

	go func(sub *pubsub.Subscription) {
		if err = sub.Receive(ctx, func(c context.Context, m *pubsub.Message) {
			log.Printf("Got message %s", string(m.Data))

			m.Ack()

			doneChan <- m.Data
		}); err != nil {
			errChan <- err
		}
	}(sub)

	log.Printf("about to send to topic %s", topic.ID())

	res := suite.client.Topic(topic.ID()).Publish(
		ctx,
		&pubsub.Message{
			Data: []byte("bryanawesome"),
		},
	)

	<-res.Ready()

	select {
	case err := <-errChan:
		log.Fatalf("failed to receive message %v", err)
	case data := <-doneChan:
		suite.Assert().Equal("bryanawesome", string(data))
		log.Println("message received successfully")

		close(doneChan)
		close(errChan)
	}

	// Remove subscription / topic
	if err := sub.Delete(ctx); err != nil {
		log.Fatalf("failed to delete sub %s", sub.ID())
	}

	if err := topic.Delete(ctx); err != nil {
		log.Fatalf("failed to delete topic %s", topic.ID())
	}
}
func TestPubsuberTestSuite(t *testing.T) {
	suite.Run(t, new(PubsuberTestSuite))
}
