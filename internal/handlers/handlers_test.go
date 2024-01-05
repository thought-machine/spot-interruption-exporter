package handlers

import (
	"sync"
	"testing"

	gcppubsub "cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/suite"
	"github.com/thought-machine/spot-interruption-exporter/internal/cache"
	"github.com/thought-machine/spot-interruption-exporter/internal/handlers/test_data"
	"github.com/thought-machine/spot-interruption-exporter/internal/metrics/mocks"
	"go.uber.org/zap"
)

type HandlersTestSuite struct {
	suite.Suite
	mockMetrics *mocks.Client
	l           *zap.SugaredLogger
}

var (
	mockInterruptionMessage = &gcppubsub.Message{
		ID:   "12345",
		Data: test_data.InterruptionEventJSONFile,
	}
	mockCreationMessage = &gcppubsub.Message{
		ID:   "12345",
		Data: test_data.CreationEventJSONFile,
	}
)

func (suite *HandlersTestSuite) SetupSuite() {
	suite.mockMetrics = mocks.NewClient(suite.T())
	l, err := zap.NewDevelopment()
	suite.NoError(err)
	suite.l = l.Sugar()
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (suite *HandlersTestSuite) TestHandleInterruptionEvents() {
	suite.mockMetrics.EXPECT().IncreaseInterruptionEventCounter("fake-cluster").Times(1)
	initialInstances := map[string]string{
		"projects/mock-project/zones/europe-west1-c/instances/mock-instance-spot-3706-5b909138-nr65": "fake-cluster",
	}
	instanceToClusterMappings := cache.NewCacheWithTTLFrom(cache.NoExpiration, initialInstances)
	interruptions := make(chan *gcppubsub.Message)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go HandleInterruptionEvents(interruptions, instanceToClusterMappings, suite.mockMetrics, suite.l, wg)
	interruptions <- mockInterruptionMessage
	interruptions <- mockInterruptionMessage
	close(interruptions)
	wg.Wait()
}

func (suite *HandlersTestSuite) TestHandleCreationEvents() {
	fakeClusterName := "fake-cluster"
	fakeInstanceName := "fake-instance"
	initialInstances := map[string]string{
		fakeInstanceName: fakeClusterName,
	}
	instanceToClusterMappings := cache.NewCacheWithTTLFrom(cache.NoExpiration, initialInstances)
	additions := make(chan *gcppubsub.Message)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go HandleCreationEvents(additions, instanceToClusterMappings, suite.l, wg)
	additions <- mockCreationMessage
	close(additions)
	wg.Wait()
	resourceName := "projects/mock-project/zones/europe-west1-c/instances/fake-resource"

	cluster, err := instanceToClusterMappings.Get(resourceName)
	suite.NoError(err)
	suite.Equal(fakeClusterName, cluster)

	cluster, err = instanceToClusterMappings.Get(fakeInstanceName)
	suite.NoError(err)
	suite.Equal(fakeClusterName, cluster)
}

func (suite *HandlersTestSuite) TestMessageToInstanceInterruptionEvent() {
	event, err := messageToInstanceInterruptionEvent(mockInterruptionMessage)
	suite.NoError(err)
	suite.Equal("projects/mock-project/zones/europe-west1-c/instances/mock-instance-spot-3706-5b909138-nr65", event.ResourceID)
	suite.Equal("12345", event.MessageID)
}

func (suite *HandlersTestSuite) TestMessageToInstanceCreationEvent() {
	event, err := messageToInstanceCreationEvent(mockCreationMessage)
	suite.NoError(err)
	suite.Equal("projects/mock-project/zones/europe-west1-c/instances/fake-resource", event.ResourceID)
	suite.Equal("fake-cluster", event.ClusterName)
	suite.Equal("12345", event.MessageID)
}
