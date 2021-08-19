package pubsuber

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

type DPPubsuber interface {
	Client() *pubsub.Client
	CreateInquiryTopic(ctx context.Context, inquiryUuid string) (*pubsub.Topic, error)
}

type DPPubsub struct {
	c *pubsub.Client
}

func New(c *pubsub.Client) *DPPubsub {
	return &DPPubsub{
		c: c,
	}
}

func (r *DPPubsub) Client() *pubsub.Client {
	return r.c
}

// CreateInquiryTopic when male user starts an inquiry, a new inquiry topic
// will be created. It is used to receive FCM messages when female has picked
// up the inquiry.
func (r *DPPubsub) CreateInquiryTopic(ctx context.Context, inquiryUuid string) (*pubsub.Topic, error) {
	curts := time.Now().UTC().Unix()
	topicName := fmt.Sprintf("%s_%s_%d", "inquiry", inquiryUuid, curts)

	topic, err := r.c.CreateTopic(ctx, topicName)

	if err != nil {
		return nil, err
	}

	return topic, err
}
