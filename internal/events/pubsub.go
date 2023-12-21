package events

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"context"
	"go.uber.org/zap"
)

type subscription struct {
	t   *gcppubsub.Subscription
	log *zap.SugaredLogger
}

// Subscription provides a wrapper around a specific pubsub subscription
type Subscription interface {
	// Receive sends outstanding pubsub messages to event channel. It blocks until ctx is done, or the pubsub returns a non-retryable error.
	Receive(ctx context.Context, event chan<- *gcppubsub.Message)
}

func (s *subscription) Receive(ctx context.Context, event chan<- *gcppubsub.Message) {
	err := s.t.Receive(ctx, func(ctx context.Context, m *gcppubsub.Message) {
		m.Ack()
		event <- m
	})
	close(event)
	if err != nil {
		s.log.Fatalf("unexpected interruption while handling messages from pubsub topic %s", err.Error())
	}
}

// PubSubNotifierInput defines all required fields to create a PubSubNotifier
type PubSubNotifierInput struct {
	Logger           *zap.SugaredLogger
	ProjectID        string
	SubscriptionName string
}

// NewPubSubNotifier returns a client that provides wrappers around the specified pubsub subscription
func NewPubSubNotifier(ctx context.Context, input *PubSubNotifierInput) (Subscription, error) {
	client, err := gcppubsub.NewClient(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	return &subscription{
		t:   client.Subscription(input.SubscriptionName),
		log: input.Logger.With("subscription_name", input.SubscriptionName),
	}, nil
}
