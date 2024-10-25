package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// 模拟缓存系统
type Cache4 struct {
	data      map[string]string
	expire    map[string]time.Time
	mutex     sync.RWMutex
	expireDur time.Duration
}

// 初始化缓存
func NewCache4(expireDur time.Duration) *Cache4 {
	return &Cache4{
		data:      make(map[string]string),
		expire:    make(map[string]time.Time),
		expireDur: expireDur,
	}
}

// 获取缓存数据，如果缓存过期则返回空字符串
func (c *Cache4) Get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 检查缓存是否存在以及是否过期
	if value, found := c.data[key]; found {
		if time.Now().Before(c.expire[key]) {
			return value, true
		}
		// 如果过期，删除缓存
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.data, key)
		delete(c.expire, key)
		c.mutex.Unlock()
		c.mutex.RLock()
	}
	return "", false
}

// 设置缓存数据
func (c *Cache4) Set(key string, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
	c.expire[key] = time.Now().Add(c.expireDur)
}

// 模拟数据库查询
func queryFromDB4(key string) string {
	fmt.Printf("Querying DB for key: %s\n", key)
	time.Sleep(100 * time.Millisecond) // 模拟数据库延迟
	return "Data for " + key
}

// 模拟高并发下缓存雪崩的情况
func SimulateCacheAvalanche() {
	cache := NewCache4(2 * time.Second) // 设置缓存过期时间为 2 秒
	var wg sync.WaitGroup

	// 模拟多次请求
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", rand.Intn(5)) // 随机生成 5 个 key

			// 检查缓存
			value, found := cache.Get(key)
			if !found {
				// 如果缓存未命中，查询数据库并更新缓存
				fmt.Printf("Cache miss for key: %s\n", key)
				value = queryFromDB4(key)
				cache.Set(key, value)
			} else {
				fmt.Printf("Cache hit for key: %s\n", key)
			}
		}(i)
	}

	wg.Wait()
}
