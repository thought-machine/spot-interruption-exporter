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

type Subscription interface {
	Receive(ctx context.Context, event chan<- *gcppubsub.Message)
}

func (s *subscription) Receive(ctx context.Context, event chan<- *gcppubsub.Message) {
	err := s.t.Receive(ctx, func(ctx context.Context, m *gcppubsub.Message) {
		m.Ack()
		event <- m
	})
	if err != nil {
		s.log.Error(err)
	}
	close(event)
}

// NewPubSubNotifierInput defines all required fields to create a PubSubNotifier
type NewPubSubNotifierInput struct {
	Logger           *zap.SugaredLogger
	ProjectID        string
	SubscriptionName string
}

// NewPubSubNotifier receives messages from the spot-instances-preemption-events subscription
func NewPubSubNotifier(ctx context.Context, input *NewPubSubNotifierInput) (Subscription, error) {
	client, err := gcppubsub.NewClient(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	return &subscription{
		t:   client.Subscription(input.SubscriptionName),
		log: input.Logger,
	}, nil
}
