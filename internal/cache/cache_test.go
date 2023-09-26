package cache

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type CacheTestSuite struct {
	suite.Suite
	sut Cache
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (suite *CacheTestSuite) TestNewCacheWithTTL() {
	c := NewCacheWithTTL(time.Second)
	c.Insert("key")
	suite.True(c.Exists("key"))
	time.Sleep(time.Millisecond * 1100)
	suite.False(c.Exists("key"))
}
