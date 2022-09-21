package main

import (
	"context"
	"fmt"
	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

type Keeper interface {
	Send(ctx context.Context, data []byte) *pubsub.PublishResult
	Stop()
}

type GooglePubSub struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewKeeper(projectId, topicName, jwtPath string,
	publishSetting *pubsub.PublishSettings) (Keeper, error) {
	if projectId == "" || topicName == "" {
		return nil, fmt.Errorf("[err] NewKeeper empty params")
	}

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectId)

	fmt.Printf("[pubsub-go] pubsub client: %+v\n", client)

	if err != nil {
		return nil, errors.Wrap(err, "[err] pubsub client")
	}

	topic := client.TopicInProject(topicName, projectId)

	fmt.Printf("[pubsub-go] pubsub topic: %+v\n", topic)

	if publishSetting != nil {
		topic.PublishSettings = *publishSetting
	} else {
		topic.PublishSettings = pubsub.DefaultPublishSettings
	}

	pubs := &GooglePubSub{client: client, topic: topic}
	return Keeper(pubs), nil
}

func (gps *GooglePubSub) Send(ctx context.Context, data []byte) *pubsub.PublishResult {
	if len(data) == 0 {
		return nil
	}
	return gps.topic.Publish(ctx, &pubsub.Message{Data: data})
}

func (gps *GooglePubSub) Stop() {
	gps.topic.Stop()
}
