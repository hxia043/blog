### 1. golang 两个协成交替打印1-100的奇数偶数

```
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	numberChan := make(chan int)

	wg.Add(1)
	go func(numberChan chan int) {
		defer wg.Done()
		for {
			number, ok := <-numberChan
			if ok {
				if number <= 100 {
					fmt.Println("id:", 1, number)
					numberChan <- number + 1
				} else {
					close(numberChan)
					break
				}
			} else {
				break
			}
		}
	}(numberChan)

	wg.Add(1)
	go func(numberChan chan int) {
		defer wg.Done()
		numberChan <- 1
		for {
			number, ok := <-numberChan
			if ok {
				if number <= 100 {
					fmt.Println("id:", 2, number)
					numberChan <- number + 1
				} else {
					close(numberChan)
					break
				}

			} else {
				break
			}
		}
	}(numberChan)

	wg.Wait()
}
```

这里我们模拟辩论机制，设两个协程，奇数协程负责输出奇数，偶数协程负责输出偶数。它们之间通过通道顺序沟通。

1. 这里，我们首先将通道里装 flag，flag 用于表明到谁输出，我们在 for 循环内打印循环变量输出。后面发现通道中的变量可以直接作为输出结果，于是省略了 for 循环体更加清晰。
2. 我们想着解耦，设置 finialzer 负责关闭通道，后来发现关闭通道还是由协程来负责比较好，只是这里关闭的时机要设置好，避免死锁。做到了简洁，清晰。
3. 我们的直觉是用除 2 取余判断奇偶数的方式来输出，后面发现由于是交替输出，我们只需要设置初始协程是奇数协程（输出 1），然后每个协程轮流递增，就可以实现交替打印奇数和偶数。这要比示例 [golang 两个协成交替打印1-100的奇数偶数](https://www.jishuchi.com/read/go-interview/3439) 更简介。我们得到了初始条件，让协程轮转起来，那么之后就是确定终止条件了。
4. 刚开始确定终止条件是想着按流程来，后面发现这会陷入一个很繁琐的流程里去。于是，转换思路，去看每一步协程需要怎么判断，怎么处理，如果通道关闭了怎么办，顺着这个思路理下去，我们才算是理通了。

总的来说，这题很不错，适合多做几次。



### 2. golang互斥锁的两种实现

两种实现互斥锁的方式，一种是 sync.Mutex，一种是通道。

#### 2.1 sync.Mutex

```
var global int
var mu sync.Mutex
var wg sync.WaitGroup

func main() {
	for i := 0; i < runtime.GOMAXPROCS(runtime.NumCPU()); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				global++
				mu.Unlock()

				if global < 20 {
					fmt.Println(global)
				} else {
					break
				}
			}
		}()
	}

	wg.Wait()
}
```

#### 2.2 通道

```
func main() {
	var numberChan = make(chan *int)
	for i := 0; i < runtime.GOMAXPROCS(runtime.NumCPU()); i++ {
		wg.Add(1)
		go func(numberChan chan *int, id int) {
			defer wg.Done()
			for {
				global, ok := <-numberChan
				if !ok {
					fmt.Printf("id: %d exit\n", id)
					break
				} else {
					if *global < 30 {
						fmt.Println("id:", id, *global)
						*global++
						numberChan <- global
					} else {
						close(numberChan)
						fmt.Printf("id: %d close channel\n", id)
						break
					}
				}
			}
		}(numberChan, i)
	}

	wg.Add(1)
	go func(global *int, numberChan chan *int) {
		defer wg.Done()
		numberChan <- global
	}(&global, numberChan)

	wg.Wait()

	fmt.Println(global)
}
```

[golang互斥锁的两种实现](https://www.jishuchi.com/read/go-interview/3439) 实现了有缓冲通道的同步，通过设置缓冲长度为 1 保证每次只有一个通道对 number 进行操作。

我们这里利用无缓冲通道给出了协程同步的实现，都是利用阻塞。

两种方式最好都要掌握。
