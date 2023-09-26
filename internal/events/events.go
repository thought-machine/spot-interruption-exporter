// Package events defines an interface for receiving interruption events, along with cloud-provider specific implementations
package events

import (
	"context"
)

// Notifier is responsible for sending InterruptionEvents to event channel whenever an instance is interrupted.
// There should generally be one Notifier implementation per cloud provider
type Notifier interface {
	Receive(ctx context.Context, event chan<- InterruptionEvent)
}

// InterruptionEvent represents a generic event when an instance is interrupted, regardless of cloud provider
type InterruptionEvent struct {
	MessageID  string
	ResourceID string
}
