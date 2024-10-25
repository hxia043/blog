package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom"
)

// 定义缓存结构体，包含数据值和过期时间
type CacheItem3 struct {
	Value      string
	Expiration int64
}

type Cache3 struct {
	data map[string]CacheItem3
	mu   sync.RWMutex
}

// 创建新的缓存
func NewCache3() *Cache3 {
	return &Cache3{
		data: make(map[string]CacheItem3),
	}
}

// 获取缓存数据
func (c *Cache3) Get3(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.data[key]
	if !found {
		return "", false
	}

	// 如果数据过期，则返回未命中
	if time.Now().UnixNano() > item.Expiration {
		return "", false
	}
	return item.Value, true
}

// 设置缓存数据
func (c *Cache3) Set3(key string, value string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置缓存和过期时间
	c.data[key] = CacheItem3{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}

// 模拟从数据库获取数据
func queryFromDB3(key string) string {
	fmt.Printf("Querying from DB for key: %s\n", key)
	time.Sleep(100 * time.Millisecond) // 模拟数据库延迟
	return "Data from DB for " + key
}

// 布隆过滤器 + 缓存的实现
type CacheWithBloom struct {
	Cache3
	Bf *bloom.BloomFilter
}

// 创建带有布隆过滤器的缓存
func NewCacheWithBloom() *CacheWithBloom {
	// 初始化布隆过滤器，假设有 1000 个元素，错误率为 0.01
	bf := bloom.New(1000*20, 5) // 5 个哈希函数
	return &CacheWithBloom{
		Cache3: *NewCache3(),
		Bf:     bf,
	}
}

// 高并发获取数据，使用布隆过滤器防止缓存穿透
func (c *CacheWithBloom) GetOrSet(key string, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	// 使用布隆过滤器判断 key 是否有可能存在
	if !c.Bf.TestString(key) {
		fmt.Printf("Bloom filter: Key %s does not exist, skipping DB query\n", key)
		return
	}

	// 先从缓存获取数据
	value, found := c.Get3(key)
	if found {
		fmt.Printf("Cache hit for key: %s\n", key)
		return
	}

	// 如果缓存未命中，模拟从数据库加载数据
	fmt.Printf("Cache miss for key: %s, querying from DB\n", key)
	value = queryFromDB(key)

	// 更新缓存
	c.Set3(key, value, duration)

	// 同时将 key 添加到布隆过滤器中
	c.Bf.AddString(key)

	fmt.Printf("Updated cache for key: %s and added to Bloom filter\n", key)
}
