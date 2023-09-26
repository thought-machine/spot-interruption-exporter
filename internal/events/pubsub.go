package events

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"context"
	"github.com/googleapis/google-cloudevents-go/cloud/auditdata"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type notifier struct {
	t   Subscription
	log *zap.SugaredLogger
}

type Subscription interface {
	Receive(ctx context.Context, f func(context.Context, *gcppubsub.Message)) error
}

func (n *notifier) Receive(ctx context.Context, event chan<- InterruptionEvent) {
	err := n.t.Receive(ctx, func(ctx context.Context, m *gcppubsub.Message) {
		m.Ack()
		e, err := formatInterruptionEventFromPubSubMessage(m)
		if err != nil {
			n.log.Errorf("failed to handle event: %s", err.Error())
			return
		}
		event <- e
	})
	if err != nil {
		n.log.Error(err)
	}
	close(event)
}

func formatInterruptionEventFromPubSubMessage(m *gcppubsub.Message) (InterruptionEvent, error) {
	entry := auditdata.LogEntryData{}
	err := protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(m.Data, &entry)
	if err != nil {
		return InterruptionEvent{}, err
	}
	return InterruptionEvent{
		MessageID:  m.ID,
		ResourceID: entry.ProtoPayload.ResourceName,
	}, nil
}

// NewPubSubNotifierInput defines all required fields to create a PubSubNotifier
type NewPubSubNotifierInput struct {
	Logger           *zap.SugaredLogger
	ProjectID        string
	SubscriptionName string
}

// NewPubSubNotifier receives messages from the spot-instances-preemption-events subscription
func NewPubSubNotifier(ctx context.Context, input *NewPubSubNotifierInput) (Notifier, error) {
	client, err := gcppubsub.NewClient(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	return &notifier{
		t:   client.Subscription(input.SubscriptionName),
		log: input.Logger,
	}, nil
}
