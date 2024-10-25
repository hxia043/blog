// The demo is for cache breakdown

package cache

import (
	"fmt"
	"sync"
	"time"
)

type CacheItem struct {
	Value      string
	Expiration int64
}

type Cache struct {
	data   map[string]CacheItem
	mu     sync.RWMutex
	dbLock sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]CacheItem),
	}
}

func (c *Cache) Get(key string) (string, bool) {
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

func (c *Cache) Set(key string, value string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}

func queryFromDB(key string) string {
	fmt.Printf("Querying from DB for key: %s\n", key)
	time.Sleep(100 * time.Millisecond) // 模拟数据库延迟
	return "Data from DB for " + key
}

func (c *Cache) GetOrSet(key string, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	retryTimes := 5
	for i := 0; i < retryTimes; i++ {
		value, found := c.Get(key)
		if found {
			fmt.Printf("Cache hit for [%s:%s]\n", key, value)
			return
		}
		fmt.Printf("Cache miss for key: %s\n", key)
	}

	// add lock from caller
	c.dbLock.Lock()

	value, found := c.Get(key)
	if found {
		fmt.Printf("Cache hit for [%s:%s]\n", key, value)
		c.dbLock.Unlock()
		return
	}

	value = queryFromDB(key)
	c.Set(key, value, duration)
	fmt.Printf("Updated cache for key: %s\n", key)
	c.dbLock.Unlock()
}
