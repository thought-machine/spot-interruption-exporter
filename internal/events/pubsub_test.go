package events

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PubSubTestSuite struct {
	suite.Suite
}

func TestPubSubTestSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}

func (suite *PubSubTestSuite) TestFormatInterruptionEventFromPubSubMessage() {
	mockMessage := &gcppubsub.Message{
		ID: "12345",
		Data: []byte(`
{
  "protoPayload": {
    "@type": "type.googleapis.com/google.cloud.audit.AuditLog",
    "status": {
      "message": "Instance was preempted."
    },
    "authenticationInfo": {
      "principalEmail": "system@google.com"
    },
    "serviceName": "compute.googleapis.com",
    "methodName": "compute.instances.preempted",
    "resourceName": "projects/mock-project/zones/europe-west1-c/instances/mock-instance-spot-3706-5b909138-nr65",
    "request": {
      "@type": "type.googleapis.com/compute.instances.preempted"
    }
  },
  "insertId": "qnwer3e38dfz",
  "resource": {
    "type": "gce_instance",
    "labels": {
      "instance_id": "184448819...",
      "project_id": "mock-project",
      "zone": "europe-west1-c"
    }
  },
  "timestamp": "2023-09-16T10:42:31.325309Z",
  "severity": "INFO",
  "logName": "projects/mock-project/logs/cloudaudit.googleapis.com%2Fsystem_event",
  "operation": {
    "id": "systemevent-1694860946116....",
    "producer": "compute.instances.preempted",
    "first": true,
    "last": true
  },
  "receiveTimestamp": "2023-09-16T10:42:31.782066320Z"
}`),
	}
	event, err := formatInterruptionEventFromPubSubMessage(mockMessage)
	suite.NoError(err)
	suite.Equal("projects/mock-project/zones/europe-west1-c/instances/mock-instance-spot-3706-5b909138-nr65", event.ResourceID)
	suite.Equal("12345", event.MessageID)
}
