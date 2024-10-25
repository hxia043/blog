### 逃逸分析

逃逸分析是 Go 语言中的一项重要优化技术，可以帮助程序减少内存分配和垃圾回收的开销，从而提高程序的性能。下面是一道涉及逃逸分析的面试题及其详解。

```
func foo() *int {
    x := 1
    return &x
}

func main() {
    p := foo()
    fmt.Println(*p)
}
```

请问上面的代码中，变量x是否会发生逃逸？


会逃逸，变量 x 在函数外有引用，编译器在逃逸分析阶段会将该变量申请在堆中分配。

### 延迟语句

defer语句是Go语言中的一项重要特性，可以用于在函数返回前执行一些清理或收尾工作，例如释放资源、关闭连接等。下面是一道涉及defer语句的面试题及其详解。

```
func main() {
    defer func() {
        fmt.Println("defer 1")
    }()
    defer func() {
        fmt.Println("defer 2")
    }()
    fmt.Println("main")
}
```

延迟调用是按照先进后出的顺序执行，以保证依赖顺利执行。本例中的输出为：
```
main
defer 2
defer 1
```

### Map

Go语言中的map是一种非常有用的数据结构，可以用于存储键值对。下面是一道涉及map的面试题及其详解。

```
func main() {
    m := make(map[int]string)
    m[1] = "a"
    m[2] = "b"
    fmt.Println(m[1], m[2])
    delete(m, 2)
    fmt.Println(m[2])
}
```

map 在获取不到元素时，会返回类型的零值，这里 string 的零值是 ""。输出：
```
a, b

```

### 通道

Go语言中的通道（channel）是一种非常有用的特性，用于在不同的goroutine之间传递数据。下面是一道涉及通道的面试题及其详解。

```
func main() {
    ch := make(chan int)
    go func() {
        ch <- 1
        ch <- 2
        ch <- 3
        close(ch)
    }()
    for {
        n, ok := <-ch
        if !ok {
            break
        }
        fmt.Println(n)
    }
    fmt.Println("done")
}
```

请问上面的代码中，输出的结果是什么？


这里首先确定通道是无缓冲的，往通道写数据时，如果没有读则写数据将阻塞。比如往通道写 int 1： `ch <-1` 时，`n, ok := <-ch` 未读取数据，则通道将阻塞在写 1，所以这里的读写数据是按顺序执行的。输出：
```
1
2
3
done
```

### 接口

Go语言中的接口（interface）是一种非常重要的特性，用于定义一组方法。下面是一道涉及接口的面试题及其详解。

```
type Animal interface {
    Speak() string
}

type Dog struct{}

func (d *Dog) Speak() string {
    return "Woof!"
}

type Cat struct{}

func (c *Cat) Speak() string {
    return "Meow!"
}

func main() {
    animals := []Animal{&Dog{}, &Cat{}}
    for _, animal := range animals {
        fmt.Println(animal.Speak())
    }
}
```

请问上面的代码中，输出的结果是什么？

对象 `Dog`，`Cat` 分别实现了接口，这里接口调用 `Speak` 实际调用的是不同对象的 `Speak` 方法。输出：
```
Woof!
Meow!
```

### 锁

在 Go 语言的面试中，锁机制是一个常见的考察点，特别是对并发编程和线程安全的理解。下面是一些常见的与锁相关的面试题及其解答思路。

#### 1. **什么是 Mutex？如何使用？**

**问题：**
- 什么是 `Mutex`？
- 如何在 Go 中使用 `Mutex`？

**解答：**
- `Mutex` 是 Go 提供的一种互斥锁，用于确保同一时间只有一个 goroutine 可以访问共享资源，防止数据竞争。
- 使用 `sync.Mutex` 的 `Lock` 方法来获取锁，使用 `Unlock` 方法来释放锁。例如：
  
  ```go
  var mu sync.Mutex
  mu.Lock()
  // critical section
  mu.Unlock()
  ```

### 2. **什么是 RWMutex？它与 Mutex 有什么区别？**

**问题：**
- 解释 `RWMutex` 及其用途。
- `RWMutex` 和 `Mutex` 有什么区别？

**解答：**
- `RWMutex` 是一种读写互斥锁，提供了两种锁定方式：读锁（`RLock`）和写锁（`Lock`）。
- 多个读操作可以并发进行，但写操作是互斥的。`RWMutex` 适用于读多写少的场景。
- `Mutex` 是一种普通的互斥锁，读写操作都是互斥的。而 `RWMutex` 允许多个读操作并发，但写操作与其他读写操作互斥。

### 3. **什么是数据竞争？如何避免数据竞争？**

**问题：**
- 什么是数据竞争？
- 如何在 Go 中避免数据竞争？

**解答：**
- 数据竞争是指两个或多个 goroutine 同时访问同一块内存，其中至少有一个是写操作，且操作未被正确同步。
- 避免数据竞争的常用方法是使用 `Mutex` 或 `RWMutex` 来同步对共享资源的访问，确保同一时间只有一个 goroutine 进行写操作。

### 4. **如何实现一个线程安全的计数器？**

**问题：**
- 实现一个线程安全的计数器，要求支持并发递增和读取操作。

**解答：**
- 可以使用 `sync.Mutex` 或 `sync.RWMutex` 来实现线程安全的计数器。
  
  ```go
  type Counter struct {
      mu    sync.Mutex
      count int
  }

  func (c *Counter) Increment() {
      c.mu.Lock()
      defer c.mu.Unlock()
      c.count++
  }

  func (c *Counter) Value() int {
      c.mu.Lock()
      defer c.mu.Unlock()
      return c.count
  }
  ```

### 5. **解释 Go 中的死锁？如何避免？**

**问题：**
- 什么是死锁？
- 在 Go 中，如何避免死锁？

**解答：**
- 死锁是指两个或多个 goroutine 相互等待对方释放资源，导致程序无法继续执行。
- 避免死锁的方法包括：
  - 保持获取锁的顺序一致。
  - 尽量减少持有锁的时间。
  - 避免嵌套锁。
  - 使用 `defer` 语句确保锁被正确释放。

### 6. **Go 中的 `sync.Cond` 是什么？如何使用？**

**问题：**
- 解释 `sync.Cond` 的用途和工作原理。
- 给出一个使用 `sync.Cond` 的例子。

**解答：**
- `sync.Cond` 是 Go 中的一种条件变量，允许 goroutine 等待或广播某个条件的变化。
- 常用于协调多个 goroutine 之间的执行顺序。
  
  示例：
  
  ```go
  var mu sync.Mutex
  var cond = sync.NewCond(&mu)
  ready := false

  func waitForCondition() {
      mu.Lock()
      for !ready {
          cond.Wait()
      }
      mu.Unlock()
  }

  func signalCondition() {
      mu.Lock()
      ready = true
      cond.Signal()  // or cond.Broadcast() for waking up all
      mu.Unlock()
  }
  ```

### 7. **如何避免并发编程中的活锁？**

**问题：**
- 什么是活锁？
- 在 Go 中，如何避免活锁？

**解答：**
- 活锁是指系统中两个或多个 goroutine 持续地进行状态变化，但却无法前进到下一个状态（通常是因为每个 goroutine 都试图“友好地”让出资源）。
- 避免活锁的方法包括：引入随机性、确保在争用资源时有明确的优先级。

### 8. **解释 `sync.Once` 的用法？**

**问题：**
- 什么是 `sync.Once`？
- 在什么情况下使用 `sync.Once`？

**解答：**
- `sync.Once` 用于确保某些初始化代码只执行一次，通常用于单例模式的实现。
- 示例：

  ```go
  var once sync.Once
  func initialize() {
      once.Do(func() {
          // initialization code
      })
  }
  ```

### 9. **什么是原子操作？Go 中如何实现原子操作？**

**问题：**
- 什么是原子操作？
- 如何在 Go 中实现原子操作？

**解答：**
- 原子操作是指不可分割的操作，确保在并发环境下操作的完整性。
- Go 提供了 `sync/atomic` 包来实现原子操作，如 `AddInt32`、`LoadInt32`、`CompareAndSwapInt32` 等。

### 10. **在 Go 中如何优雅地关闭一个 goroutine？**

**问题：**
- 在 Go 中，如何优雅地关闭一个 goroutine？

**解答：**
- 使用 `context` 包或 `channel` 进行 goroutine 的取消通知。
- 示例：

  ```go
  ctx, cancel := context.WithCancel(context.Background())

  go func() {
      select {
      case <-ctx.Done():
          // cleanup and exit
      }
  }()

  // 当需要关闭 goroutine 时调用 cancel
  cancel()
  ```

这些问题覆盖了 Go 中锁的常见概念和使用场景。了解这些问题及其解答，可以帮助你在面试中更好地展示你的并发编程能力。


### 11. 线程，进程，协程的区别？

### 

```
这个过程怎么看？ M 是需要是创建还是初始时创建？

Go 程序启动后，会给每个逻辑核心分配一个 P（Logical Processor）；同时，会给每个 P 分配一个 M（Machine，表示内核线程），这些内核线程仍然由 OS scheduler 来调度。
```


```
G、P、M 都说完了，还有两个比较重要的组件没有提到：全局可运行队列（GRQ）和本地可运行队列（LRQ）。LRQ 存储本地（也就是具体的 P）的可运行 goroutine，GRQ 存储全局的可运行 goroutine，这些 goroutine 还没有分配到具体的 P。
```

### 12. goroutine 的调度时机？

```
由于 Go 语言是协作式的调度，不会像线程那样，在时间片用完后，由 CPU 中断任务强行将其调度走。对于 Go 语言中运行时间过长的 goroutine，Go scheduler 有一个后台线程在持续监控，一旦发现 goroutine 运行超过 10 ms，会设置 goroutine 的“抢占标志位”，之后调度器会处理。但是设置标志位的时机只有在函数“序言”部分，对于没有函数调用的就没有办法了。
```

怎么理解 `但是设置标志位的时机只有在函数“序言”部分，对于没有函数调用的就没有办法了` 这句话？

```
type g struct {
// 指向全局队列里下一个 g

    schedlink      guintptr     // 这个是干嘛的？
}
```


```
能构造这样的场景吗？


一个 M 只有绑定 P 才能执行 goroutine，当 M 被阻塞时，整个 P 会被传递给其他 M ，或者说整个 P 被接管。
```

这句话说的挺好的：
GPM 三足鼎力，共同成就 Go scheduler。G 需要在 M 上才能运行，M 依赖 P 提供的资源，P 则持有待运行的 G


系统加载可执行文件的过程：
```
从磁盘上读取可执行文件，加载到内存

创建进程和主线程

为主线程分配栈空间

把由用户在命令行输入的参数拷贝到主线程的栈

把主线程放入操作系统的运行队列等待被调度
```

```
第一行，将 gp.sched.g 保存到 DX 寄存器；第二行，我们见得已经比较多了， get_tls 将 tls 保存到 CX 寄存器，再将 gp.sched.g 放到 tls[0] 处。这样，当下次再调用 get_tls 时，取出的就是 gp，而不再是 g0，这一行完成从 g0 栈切换到 gp。
```

怎么知道 runtime.main 的地址的？
```
最终，调度器完成了这个值得铭记的时刻，从 g0 转到 gp，开始执行 runtime.main 函数。

用一张流程图总结一下从 g0 切换到 main goroutine 的过程：
```

开辟一个栈，在栈内运行 schedule。

```
什么时候 m.p 换成 1 了？
```

### 13. 关于 GC 需要深入了解的

这段代码片段和注释来自 Go runtime 的源码，涉及到 Go 的垃圾回收器（Garbage Collector, GC）以及 `M` 结构体的内存管理。以下是代码的逐步解析以及其意义：

### 背景

Go runtime 使用了 **M**、**P** 和 **G** 模型进行 goroutine 的调度。其中，`M`（Machine）代表操作系统的线程，`G`（Goroutine）代表一个 goroutine，`P`（Processor）则是调度器中负责分配 `G` 给 `M` 执行的逻辑单元。为了有效管理这些 goroutine 和线程，Go runtime 需要处理内存的分配和回收，包括 **防止垃圾回收器误释放活动对象**。

在 Go runtime 的运行过程中，特别是涉及到线程和 goroutine 之间的切换时，垃圾回收器（GC）会周期性地扫描堆上和栈上的内存，以标记并清理不再使用的内存资源。但是，有时某些数据对象并不会直接在栈或堆上明确可见（例如，当它们保存在寄存器或线程局部存储中），这时需要一些特殊处理来确保这些对象不会被错误地回收。

### 代码解释

```go
// Add to allm so garbage collector doesn't free g->m
// when it is just in a register or thread-local storage.
mp.alllink = allm
```

1. **`mp.alllink = allm`**：  
   这行代码的目的是将当前 `M` 结构体的指针（`mp`）链接到一个全局的 `allm` 列表中。`allm` 是一个全局的链表结构，维护了所有活跃的 `M` 实例。在 Go runtime 中，每个 `M` 都会被添加到这个全局链表中。

2. **防止垃圾回收**：  
   注释解释了为什么要将 `mp` 加入到 `allm` 中：`M` 结构体的实例（即 `mp`）可能不会明确存储在堆或栈上，而是保存在寄存器或线程局部存储（Thread Local Storage, TLS）中。在垃圾回收过程中，GC 会扫描栈和堆以标记活动对象，但它不会自动扫描寄存器或 TLS 中的数据。

   因此，如果 `M` 实例（包括它关联的 goroutine 及其他状态）只在寄存器或 TLS 中，而不在栈或堆中，GC 可能会错误地认为这个 `M` 实例不再被引用，从而释放它的内存。这将导致未定义行为，可能会引发崩溃或其他错误。

3. **`allm` 的作用**：  
   将 `M` 添加到全局的 `allm` 链表中，确保了即使 `M` 结构体的引用不在栈或堆上，GC 仍然能找到这个 `M` 的实例并标记它为“活跃的”。这样就避免了 GC 错误地回收 `M` 实例的内存。

### 小结

这段代码的关键点是确保垃圾回收器不会错误地释放当前线程（`M`）的资源，即使 `M` 结构体的引用只存在于寄存器或线程局部存储中。通过将 `M` 添加到全局的 `allm` 链表中，Go runtime 确保了所有活跃的线程和它们关联的资源在 GC 扫描时都不会被误删除。


### 内存分配器

新创建的 goroutine 是在堆上还是栈上分配的?


### 参考

- [刷起来: Go必看的进阶面试题详解](https://mp.weixin.qq.com/s/2iOkW5h7x-1wdYe51vMemw)
- [记录一次腾讯Go开发岗位面试经过](https://learnku.com/articles/51080)
- [给大家丢脸了，用了三年golang，我还是没答对这道内存泄漏题](https://learnku.com/articles/56077)
- [面试必备(背)--Go语言八股文系列](https://cloud.tencent.com/developer/article/1900359)

