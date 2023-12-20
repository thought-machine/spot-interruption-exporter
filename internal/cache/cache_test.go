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
	expirationDuration := time.Millisecond * 500
	c := NewCacheWithTTL(expirationDuration * 2)
	c.Insert("key", "")
	suite.Eventually(func() bool {
		return c.Exists("key")
	}, time.Second*5, expirationDuration)
	suite.Eventually(func() bool {
		return !c.Exists("key")
	}, time.Second*2, expirationDuration)
}

func (suite *CacheTestSuite) TestNewCacheWithTTLFrom() {
	expirationDuration := NoExpiration
	existingItemKey := "item"
	existingItemValue := "value"
	existing := map[string]string{
		existingItemKey: existingItemValue,
	}
	c := NewCacheWithTTLFrom(expirationDuration, existing)
	v, ok := c.Get(existingItemKey)
	suite.True(ok)
	suite.Equal(existingItemValue, v)
}

func (suite *CacheTestSuite) TestSetExpiration() {
	itemKey := "item"
	c := NewCacheWithTTL(NoExpiration)
	c.SetExpiration("non-existent", NoExpiration)
	c.Insert(itemKey, "")
	c.SetExpiration(itemKey, time.Nanosecond)
	suite.Eventually(func() bool {
		_, ok := c.Get(itemKey)
		return !ok
	}, time.Second*2, time.Microsecond)
}
