//go:build integration

package compute

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type ComputeTestSuite struct {
	suite.Suite
}

func TestComputeTestSuite(t *testing.T) {
	suite.Run(t, new(ComputeTestSuite))
}

func (suite *ComputeTestSuite) SetupSuite() {
}

func (suite *ComputeTestSuite) TestGetInstancesBelongingToKubernetesClusters() {
	l, err := zap.NewDevelopment()
	suite.NoError(err)
	logger := l.Sugar()
	c, err := NewClient(context.Background(), NewClientInput{
		Logger:    logger,
		ProjectID: "generic-project",
	})
	res, err := c.ListInstancesBelongingToKubernetesCluster(context.Background())
	suite.Error(err)
	suite.NotEmpty(res)
}
