package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (suite *CacheTestSuite) TestNewCacheWithTTL() {
	expirationDuration := time.Millisecond * 250
	c := NewCacheWithTTL(expirationDuration)
	c.Insert("key", "")
	suite.True(c.Exists("key"))
	suite.Eventually(func() bool {
		return !c.Exists("key")
	}, time.Second*2, time.Millisecond*100)
}

func (suite *CacheTestSuite) TestNewCacheWithTTLFrom() {
	existingItemKey := "item"
	existingItemValue := "value"
	existing := map[string]string{
		existingItemKey: existingItemValue,
	}
	c := NewCacheWithTTLFrom(NoExpiration, existing)
	v, err := c.Get(existingItemKey)
	suite.NoError(err)
	suite.Equal(existingItemValue, v)
}

func (suite *CacheTestSuite) TestSetExpiration() {
	itemKey := "item"
	c := NewCacheWithTTL(NoExpiration)

	suite.Error(c.SetExpiration("non-existent", NoExpiration))

	c.Insert(itemKey, "")
	suite.NoError(c.SetExpiration(itemKey, time.Nanosecond))
	suite.Eventually(func() bool {
		return !c.Exists(itemKey)
	}, time.Second*2, time.Microsecond)
}
