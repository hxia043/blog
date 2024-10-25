// The demo is for cache penetration

package cache

import (
	"fmt"
	"sync"
	"time"
)

type CacheItem2 struct {
	Value      string
	Expiration int64
}

type Cache2 struct {
	data   map[string]CacheItem
	mu     sync.RWMutex
	dbLock sync.RWMutex
}

func NewCache2() *Cache2 {
	return &Cache2{
		data: make(map[string]CacheItem),
	}
}

func (c *Cache2) Get2(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.data[key]
	if !found {
		return "", false
	}

	if time.Now().UnixNano() > item.Expiration {
		return "", false
	}

	return item.Value, true
}

func (c *Cache2) Set2(key string, value string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}

func queryFromDB2(key string) string {
	fmt.Printf("Querying from DB for key: %s\n", key)
	time.Sleep(100 * time.Millisecond) // 模拟数据库延迟
	return ""
}

func (c *Cache2) GetOrSet2(key string, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	retryTimes := 5
	for i := 0; i < retryTimes; i++ {
		value, found := c.Get2(key)
		if found {
			fmt.Printf("Cache hit for [%s:%s]\n", key, value)
			return
		}
		fmt.Printf("Cache miss for key: %s\n", key)
	}

	// add lock from caller
	c.dbLock.Lock()
	defer c.dbLock.Unlock()

	value, found := c.Get2(key)
	if found {
		fmt.Printf("Cache hit for [%s:%s]\n", key, value)
		return
	}

	if value = queryFromDB2(key); value == "" {
		fmt.Printf("no key: %s found\n", key)
		c.Set2(key, value, duration)
		fmt.Printf("Updated cache for key: %s with value %s\n", key, value)
		return
	}

	c.Set2(key, value, duration)
	fmt.Printf("Updated cache for key: %s\n", key)
}
