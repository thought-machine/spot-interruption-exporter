// Package handlers contains functions for converting generic GCP PubSub messages to app-specific messages
package handlers

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"github.com/googleapis/google-cloudevents-go/cloud/auditdata"
	"github.com/thought-machine/spot-interruption-exporter/internal/cache"
	"github.com/thought-machine/spot-interruption-exporter/internal/compute"
	"github.com/thought-machine/spot-interruption-exporter/internal/metrics"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"
	"time"
)

type instanceInterruptionEvent struct {
	MessageID  string
	ResourceID string
}

type instanceCreationEvent struct {
	MessageID   string
	ResourceID  string
	ClusterName string
}

// HandleCreationEvents reads from additions and adds the instance ID and corresponding cluster to m
func HandleCreationEvents(additions chan *gcppubsub.Message, m cache.Cache, l *zap.SugaredLogger) {
	for addition := range additions {
		a, err := messageToInstanceCreationEvent(addition)
		if err != nil {
			l.Warnf("failed to convert pubsub message to creation event: %s", err.Error())
		}
		s := l.With("message_id", a.MessageID, "resource_id", a.ResourceID, "kubernetes_cluster", a.ClusterName)
		m.Insert(a.ResourceID, a.ClusterName)
		s.Info("added")
	}
}

// HandleInterruptionEvents reads from interruptions and increases the interruption event counter of metrics accordingly
func HandleInterruptionEvents(interruptions chan *gcppubsub.Message, m cache.Cache, metrics metrics.Client, l *zap.SugaredLogger) {
	messageCache := cache.NewCacheWithTTL(time.Minute * 10)
	for interruption := range interruptions {
		e, err := messageToInstanceInterruptionEvent(interruption)
		if err != nil {
			l.Warnf("failed to convert pubsub message to interruption event: %s", err.Error())
			continue
		}
		s := l.With("message_id", e.MessageID, "resource_id", e.ResourceID)
		// this ensures we do not handle a duplicate message in the event pubsub sends it more than once
		if exists := messageCache.Exists(e.MessageID); exists {
			s.Debug("handled duplicate message")
			continue
		}
		messageCache.Insert(e.MessageID, "")
		clusterName, ok := m.Get(e.ResourceID)
		if !ok {
			s.Warnf("failed to determine cluster the instance (%s) belongs to", e.ResourceID)
			return
		}
		expireAfter := time.Second * 30
		m.SetExpiration(e.ResourceID, expireAfter)
		s.Debugf("%s will no longer be tracked after %s", e.ResourceID, expireAfter)

		s.Info("interrupted")
		metrics.IncreaseInterruptionEventCounter(clusterName)
	}
}

func messageToInstanceInterruptionEvent(m *gcppubsub.Message) (instanceInterruptionEvent, error) {
	entry := auditdata.LogEntryData{}
	err := protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(m.Data, &entry)
	if err != nil {
		return instanceInterruptionEvent{}, err
	}
	return instanceInterruptionEvent{
		MessageID:  m.ID,
		ResourceID: entry.ProtoPayload.ResourceName,
	}, nil
}

func messageToInstanceCreationEvent(m *gcppubsub.Message) (instanceCreationEvent, error) {
	entry := auditdata.LogEntryData{}
	err := protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(m.Data, &entry)
	if err != nil {
		return instanceCreationEvent{}, err
	}
	labels := entry.ProtoPayload.Request.GetFields()["labels"]
	clusterName := "undefined"
	for _, v := range labels.GetListValue().Values {
		label := v.GetStructValue().AsMap()
		labelKey := label["key"].(string)
		if strings.EqualFold(labelKey, compute.ClusterNameLabelKey) {
			clusterName = label["value"].(string)
			break
		}
	}
	return instanceCreationEvent{
		MessageID:   m.ID,
		ResourceID:  entry.ProtoPayload.ResourceName,
		ClusterName: clusterName,
	}, nil

}
