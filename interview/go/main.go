package main

import (
	"fmt"
	"go-interview/cache"
	"math/rand"
	"time"
)

func main() {
	//channel.Print()
	//channel.CSP()

	/*
		cache := cache.NewCache()
		var wg sync.WaitGroup

		// 设置热点数据，过期时间为 1 秒
		cache.Set("hotkey", "Hot Data", 1*time.Second)
		cache.Set("coldkey", "Cold Data", 1*time.Second)

		// 模拟 10 个 goroutine 并发访问热点数据
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				// 模拟在 1 秒后缓存失效，触发缓存击穿
				time.Sleep(2 * time.Second)
				cache.GetOrSet("hotkey", 5*time.Second, &wg)
			}()
		}

		wg.Wait()
	*/

	/*
		cache := cache.NewCache2()
		var wg sync.WaitGroup

		// 设置热点数据，过期时间为 1 秒
		cache.Set2("hotkey", "Hot Data", 1*time.Second)
		cache.Set2("coldkey", "Cold Data", 1*time.Second)

		// 模拟 10 个 goroutine 并发访问热点数据
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				// 模拟在 1 秒后缓存失效，触发缓存击穿
				time.Sleep(2 * time.Second)
				cache.GetOrSet2("hotdogkey", 5*time.Second, &wg)
			}()
		}

		wg.Wait()
	*/

	/*
		cache := cache.NewCacheWithBloom()
		var wg sync.WaitGroup

		// 模拟数据库中的一些 key，加入布隆过滤器中
		existingKeys := []string{"key1", "key2", "key3"}
		for _, key := range existingKeys {
			cache.Bf.AddString(key)
		}

		// 模拟并发访问缓存或数据库
		keysToRequest := []string{"key1", "key2", "key100", "key101", "key3"}

		for _, key := range keysToRequest {
			wg.Add(1)
			go cache.GetOrSet(key, 5*time.Second, &wg)
		}

		wg.Wait()
	*/

	rand.Seed(time.Now().UnixNano())
	fmt.Println("Starting cache avalanche simulation...")

	for i := 0; i < 5; i++ {
		fmt.Printf("\n--- Iteration %d ---\n", i+1)
		cache.SimulateCacheAvalanche()
		time.Sleep(3 * time.Second) // 模拟每次缓存过期后重新请求
	}
}
