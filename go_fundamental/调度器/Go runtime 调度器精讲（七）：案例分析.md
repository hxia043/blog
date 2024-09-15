# 0. 前言

前面用了六讲介绍 Go runtime 调度器，这一讲我们看一个关于调度 goroutine 的程序案例分析下调度器做了什么。需要说明的是，这个程序和抢占有关，抢占目前为止还没有介绍到，如果看不懂也没有关系，有个印象就行。

# 1. 案例 1

执行代码：
```
func gpm() {
	var x int
	for {
		x++
	}
}

func main() {
	var x int
	threads := runtime.GOMAXPROCS(0)
	for i := 0; i < threads; i++ {
		go gpm()
	}

	time.Sleep(1 * time.Second)
	fmt.Println("x = ", x)
}
```

运行程序：
```
# go run main.go 
x =  0
```
(*为什么输出 x=0 和本系列内容无关，这里直接跳过*)

Go 在 [1.14](https://go.dev/doc/go1.14#runtime) 版本引入了异步抢占机制，我们使用的是 `1.21.0` 版本的 Go，默认开启异步抢占。通过 `asyncpreemptoff` 标志可以开启/禁用异步抢占，`asyncpreemptoff=1` 表示禁用异步抢占，相应的 `asyncpreemptoff=0` 表示开启异步抢占。

## 1.1 禁用异步抢占

首先，禁用异步抢占，再次执行上述代码：
```
# GODEBUG=asyncpreemptoff=1 go run main.go

```

程序卡死，无输出。查看 CPU 使用率：
```
top - 10:08:53 up 86 days, 10:48,  0 users,  load average: 3.08, 1.29, 0.56
Tasks: 179 total,   2 running, 177 sleeping,   0 stopped,   0 zombie
%Cpu(s): 74.4 us,  0.6 sy,  0.0 ni, 25.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
MiB Mem :  20074.9 total,   4279.4 free,   3118.3 used,  12677.2 buff/cache
MiB Swap:      0.0 total,      0.0 free,      0.0 used.  16781.0 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND                                 
1014008 root      20   0 1226288    944    668 R 293.7   0.0   5:35.81 main             // main 是执行的进程
```

CPU 占用率高达 `293.7`，太高了。

为什么会出现这样的情况呢？我们可以通过 [GODEBUG=schedtrace=1000,scheddetail=1,asyncpreemptoff=1](https://go.dev/wiki/Performance) 打印程序执行的 G，P，M 信息，通过 DEBUG 输出查看调度过程中发生了什么。

当创建和线程数相等的 `goroutine` 后，线程执行 `main goroutine`。runtime（实际是 sysmon 线程，后文会讲）发现 main goroutine 运行时间过长，把它调度走，运行其它 goroutine（这是主动调度的逻辑，不属于异步抢占的范畴）。接着执行和线程数相等的 goroutine，这几个 goroutine 是永不退出的，线程会一直执行，占满逻辑核。

解决这个问题，我们改动代码如下：
```
func main() {
	var x int
	threads := runtime.GOMAXPROCS(0)
	for i := 0; i < threads; i++ {
		go gpm()
	}

	time.Sleep(1 * time.Nanosecond)
	fmt.Println("x = ", x)
}
```

因为 main goroutine 运行时间过长，被 runtime 调度走。我们把休眠时间设成 1 纳秒，不让它睡那么长。接着执行程序：
```
# GODEBUG=asyncpreemptoff=1 go run main.go 
x =  0
```

程序退出。天下武功唯快不破啊，main goroutine 直接执行完退出，不给 runtime 反应的机会。

还有其它改法吗？我们在 gpm 中加上 `time.Sleep` 函数调用：
```
func gpm() {
	var x int
	for {
		time.Sleep(1 * time.Nanosecond)
		x++
	}
}

func main() {
	var x int
	threads := runtime.GOMAXPROCS(0)
	for i := 0; i < threads; i++ {
		go gpm()
	}

	time.Sleep(1 * time.Second)
	fmt.Println("x = ", x)
}
```

运行程序：
```
# GODEBUG=asyncpreemptoff=1 go run main.go 
x =  0
```

也是正常退出。为什么加上函数调用就可以呢？这和抢占的逻辑有关，因为有了函数调用，就有机会在函数序言部分设置“抢占标志”，执行抢占 goroutine 的调度（同样的，后面会详细讲）。

要注意这里 `time.Sleep(1 * time.Nanosecond)` 加的位置，如果加在这里：
```
func gpm() {
	var x int
	time.Sleep(1 * time.Nanosecond)
	for {
		x++
	}
}
```

程序还是会卡死。

我们讨论了半天 `asyncpreemptoff=1` 禁止异步抢占的情况。是时候开启异步抢占看看输出结果了。

## 1.2 开启异步抢占

程序还是那个程序：
```
func gpm() {
	var x int
	for {
		x++
	}
}

func main() {
	var x int
	threads := runtime.GOMAXPROCS(0)
	for i := 0; i < threads; i++ {
		go gpm()
	}

	time.Sleep(1 * time.Second)
	fmt.Println("x = ", x)
}
```

开启异步抢占执行：
```
# GODEBUG=asyncpreemptoff=0 go run main.go 
x =  0
```

异步抢占就可以了，为啥异步抢占就可以了呢？异步抢占通过给线程发信号的方式，使得线程在“安全点”执行异步抢占的逻辑（后面几讲会介绍异步抢占的逻辑）。

再次改写代码如下：
```
//go:nosplit
func gpm() {
	var x int
	for {
		x++
	}
}

func main() {
	var x int
	threads := runtime.GOMAXPROCS(0)
	for i := 0; i < threads; i++ {
		go gpm()
	}

	time.Sleep(1 * time.Second)
	fmt.Println("x = ", x)
}
```

同样的执行输出：
```
# GODEBUG=asyncpreemptoff=0 go run main.go 

```

程序又卡死了...

这个程序就当思考题吧，为什么加个 `//go:nosplit` 程序就卡死了呢？

# 2. 小结

本讲不是为了凑字数，主要是为引入后续的抢占做个铺垫，下一讲会介绍运行时间过长的抢占调度。

