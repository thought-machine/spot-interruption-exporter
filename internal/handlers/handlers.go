// Package handlers contains functions for converting generic GCP PubSub messages to app-specific messages
package handlers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	gcppubsub "cloud.google.com/go/pubsub"
	"github.com/googleapis/google-cloudevents-go/cloud/auditdata"
	"github.com/thought-machine/spot-interruption-exporter/internal/cache"
	"github.com/thought-machine/spot-interruption-exporter/internal/compute"
	"github.com/thought-machine/spot-interruption-exporter/internal/metrics"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
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
func HandleCreationEvents(additions chan *gcppubsub.Message, instanceToClusterMappings cache.Cache, l *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()
	for addition := range additions {
		a, err := messageToInstanceCreationEvent(addition)
		if err != nil {
			l.Warnf("failed to convert pubsub message to creation event: %s", err.Error())
			continue
		}
		l.With("message_id", a.MessageID, "resource_id", a.ResourceID, "kubernetes_cluster", a.ClusterName).Info("added")
		instanceToClusterMappings.Insert(a.ResourceID, a.ClusterName)
	}
}

// HandleInterruptionEvents reads from interruptions and increases the interruption event counter of metrics accordingly
func HandleInterruptionEvents(interruptions chan *gcppubsub.Message, instanceToClusterMappings cache.Cache, metrics metrics.Client, l *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()
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
		clusterName, err := instanceToClusterMappings.Get(e.ResourceID)
		if err != nil {
			s.Warnf("failed to determine cluster the instance (%s) belongs to: %s", e.ResourceID, err.Error())
			continue
		}
		expireAfter := time.Second * 30
		err = instanceToClusterMappings.SetExpiration(e.ResourceID, expireAfter)
		if err != nil {
			s.Warnf("failed to remove instance from mapping of instances to clusters: %s", err.Error())
		}
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
	requestFields := entry.ProtoPayload.Request.GetFields()

	labels, ok := requestFields["labels"]
	if !ok {
		return instanceCreationEvent{}, fmt.Errorf("expected labels not found on instance creation request, operation ID: %s", entry.Operation.Id)
	}

	clusterName := "unknown"
	for _, v := range labels.GetListValue().Values {
		label := v.GetStructValue().AsMap()
		labelKey := label["key"].(string)
		if strings.EqualFold(labelKey, compute.ClusterNameLabelKey) {
			clusterName = label["value"].(string)
			break
		}
	}

	if clusterName == "unknown" {
		return instanceCreationEvent{}, fmt.Errorf("expected cluster label %s not found on instance creation request, operation ID: %s", compute.ClusterNameLabelKey, entry.Operation.Id)
	}

	responseFields := entry.ProtoPayload.Response.GetFields()
	targetLink, ok := responseFields["targetLink"]
	if !ok {
		return instanceCreationEvent{}, fmt.Errorf("expected targetLink not found in instance creation response, operation ID: %s", entry.Operation.Id)
	}
	resourceID := strings.TrimPrefix(targetLink.GetStringValue(), "https://www.googleapis.com/compute/v1/")

	return instanceCreationEvent{
		MessageID:   m.ID,
		ResourceID:  resourceID,
		ClusterName: clusterName,
	}, nil
}
