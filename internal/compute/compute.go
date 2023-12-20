package compute

import (
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"strings"
)

var (
	ClusterNameLabelKey = "goog-k8s-cluster-name"
)

type Client interface {
	// ListInstancesBelongingToKubernetesCluster returns a map of all instances (key) and their corresponding Kubernetes cluster (value)
	ListInstancesBelongingToKubernetesCluster(ctx context.Context) (map[string]string, error)
}

type client struct {
	instancesClient *compute.InstancesClient
	log             *zap.SugaredLogger
	projectID       string
}

// NewClientInput defines all required fields to create a Client
type NewClientInput struct {
	Logger    *zap.SugaredLogger
	ProjectID string
}

func (c *client) listInstancesWithFilter(ctx context.Context, filter string) (map[string]string, error) {
	instancesToCluster := make(map[string]string)
	iter := c.instancesClient.AggregatedList(ctx, &computepb.AggregatedListInstancesRequest{
		Filter:  &filter,
		Project: c.projectID,
	})
	for {
		instancesInZone, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over compute instances in %s: %w", instancesInZone.Key, err)
		}
		for _, instance := range instancesInZone.Value.Instances {
			// if the label didn't exist on the instance, it won't be in the list of instances
			resourceID := strings.TrimPrefix(instance.GetSelfLink(), "https://www.googleapis.com/compute/v1/")
			instancesToCluster[resourceID] = instance.Labels[ClusterNameLabelKey]
		}
	}
	return instancesToCluster, nil
}

func (c *client) ListInstancesBelongingToKubernetesCluster(ctx context.Context) (map[string]string, error) {
	queryFilter := `labels.goog-k8s-cluster-name:*`
	return c.listInstancesWithFilter(ctx, queryFilter)
}

func NewClient(ctx context.Context, input NewClientInput) (Client, error) {
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}
	return &client{
		instancesClient: c,
		log:             input.Logger,
		projectID:       input.ProjectID,
	}, nil
}
