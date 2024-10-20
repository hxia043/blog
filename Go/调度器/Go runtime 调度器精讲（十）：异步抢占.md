# 0. 前言

前面介绍了运行时间过长和系统调用引起的抢占，它们都属于协作式抢占。本讲会介绍基于信号的真抢占式调度。

在介绍真抢占式调度之前看下 Go 的两种抢占式调度器：

**抢占式调度器 - Go 1.2 至今**
- 基于协作的抢占式调度器 - Go 1.2 - Go 1.13  
  改进：通过编译器在函数调用时插入**抢占检查**指令，在函数调用时检查当前 Goroutine 是否发起了抢占请求，实现基于协作的抢占式调度。  
  缺陷：Goroutine 可能会因为垃圾收集和循环长时间占用资源导致程序暂停。  
- 基于信号的抢占式调度器 - Go 1.14 至今  
  改进：实现了基于信号的**真抢占式调度**。  
  缺陷 1：垃圾收集在扫描栈时会触发抢占式调度。  
  缺陷 2：抢占的时间点不够多，不能覆盖所有边缘情况。  

（*注：该段文字来源于 [抢占式调度器](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-goroutine/#%E6%8A%A2%E5%8D%A0%E5%BC%8F%E8%B0%83%E5%BA%A6%E5%99%A8)*）


协作式抢占是通过在函数调用时插入 **抢占检查** 来实现抢占的，这种抢占的问题在于，如果 goroutine 中没有函数调用，那就没有办法插入 **抢占检查**，导致无法抢占。我们看 [Go runtime 调度器精讲（七）：案例分析](https://www.cnblogs.com/xingzheanan/p/18415503) 的示例：
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

禁用异步抢占：
```
# GODEBUG=asyncpreemptoff=1 go run main.go

```

程序会卡死。这是因为在 gpm 前插入 `//go:nosplit` 会禁止函数栈扩张，协作式抢占不能在函数栈调用前插入 **抢占检查**，导致这个 goroutine 没办法被抢占。

而基于信号的真抢占式调度可以改善这个问题。

# 1. 基于信号的真抢占式调度

这里我们说的异步抢占指的就是基于信号的真抢占式调度。

异步抢占的实现在 ：
```
func preemptone(pp *p) bool {
	...
	// Request an async preemption of this P.
	if preemptMSupported && debug.asyncpreemptoff == 0 {
		pp.preempt = true                                       
		preemptM(mp)                                            // 异步抢占
	}

	return true
}
```

进入 `preemptM`：
```
func preemptM(mp *m) {
	...
	if mp.signalPending.CompareAndSwap(0, 1) {                  // 更新 signalPending
		signalM(mp, sigPreempt)                                 // signalM 给线程发信号
	}
	...
}

// signalM sends a signal to mp.
func signalM(mp *m, sig int) {
	tgkill(getpid(), int(mp.procid), sig)
}

func tgkill(tgid, tid, sig int)
```

调用 `signalM` 给线程发 sigPreempt（_SIGURG：23）信号。线程接收到该信号会做相应的处理。

## 1.1 线程处理抢占信号

线程是怎么处理操作系统发过来的 sigPreempt 信号的呢？

线程的信号处理在 [sighandler](https://github.com/golang/go/blob/master/src/runtime/signal_unix.go#L619)：
```
func sighandler(sig uint32, info *siginfo, ctxt unsafe.Pointer, gp *g) {\
    // The g executing the signal handler. This is almost always
	// mp.gsignal. See delayedSignal for an exception.
	gsignal := getg()
	mp := gsignal.m

    if sig == sigPreempt && debug.asyncpreemptoff == 0 && !delayedSignal {
		// Might be a preemption signal.
		doSigPreempt(gp, c)
		// Even if this was definitely a preemption signal, it
		// may have been coalesced with another signal, so we
		// still let it through to the application.
	}
    ...
}
```

进入 `doSigPreempt`：
```
// doSigPreempt handles a preemption signal on gp.
func doSigPreempt(gp *g, ctxt *sigctxt) {
	// Check if this G wants to be preempted and is safe to
	// preempt.
	if wantAsyncPreempt(gp) {
		if ok, newpc := isAsyncSafePoint(gp, ctxt.sigpc(), ctxt.sigsp(), ctxt.siglr()); ok {
			// Adjust the PC and inject a call to asyncPreempt.
			ctxt.pushCall(abi.FuncPCABI0(asyncPreempt), newpc)
		}
	}

	// Acknowledge the preemption.
	gp.m.preemptGen.Add(1)
	gp.m.signalPending.Store(0)
}
```

首先，`doSigPreempt` 调用 `wantAsyncPreempt` 判断是否做异步抢占：
```
// wantAsyncPreempt returns whether an asynchronous preemption is
// queued for gp.
func wantAsyncPreempt(gp *g) bool {
	// Check both the G and the P.
	return (gp.preempt || gp.m.p != 0 && gp.m.p.ptr().preempt) && readgstatus(gp)&^_Gscan == _Grunning
}
```

如果是，继续调用 [isAsyncSafePoint](https://github.com/golang/go/blob/master/src/runtime/preempt.go#L363) 判断当前执行的是不是异步安全点，线程只有执行到异步安全点才能处理异步抢占。安全点是指 Go 运行时认为可以安全地暂停或抢占一个正在运行的 Goroutine 的位置。异步抢占的安全点确保 Goroutine 在被暂停或切换时，系统的状态是稳定和一致的，不会出现数据竞争、死锁或未完成的重要计算。

如果是异步抢占的安全点。则调用 `ctxt.pushCall(abi.FuncPCABI0(asyncPreempt), newpc)` 执行 `asyncPreempt`：

```
// asyncPreempt saves all user registers and calls asyncPreempt2.
//
// When stack scanning encounters an asyncPreempt frame, it scans that
// frame and its parent frame conservatively.
//
// asyncPreempt is implemented in assembly.
func asyncPreempt()                                       

//go:nosplit
func asyncPreempt2() {                          // asyncPreempt 会调用到 asyncPreempt2
	gp := getg()
	gp.asyncSafePoint = true
	if gp.preemptStop {                         
		mcall(preemptPark)                      // 抢占类型，如果是 preemptStop 则执行 preemptPark 抢占
	} else {
		mcall(gopreempt_m)                      
	}
	gp.asyncSafePoint = false
}
```

`asyncPreempt` 调用 `asyncPreempt2` 处理 `gp.preemptStop` 和非 `gp.preemptStop` 的抢占。对于非 `gp.preemptStop` 的抢占，我们在 [Go runtime 调度器精讲（八）：运行时间过长的抢占](https://www.cnblogs.com/xingzheanan/p/18415899) 也介绍过，主要内容是将运行时间过长的 goroutine 放到全局队列中。接着线程执行调度获取下一个可运行的 goroutine。

## 1.2 案例分析

还记得在 [Go runtime 调度器精讲（七）：案例分析](https://www.cnblogs.com/xingzheanan/p/18415503) 中最后留下的思考吗？
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

# GODEBUG=asyncpreemptoff=0 go run main.go 

```

为什么开启异步抢占，程序还是会卡死？

从前面的分析结合我们的 `dlv debug` 发现，在安全点判断 `isAsyncSafePoint` 这里总是返回 false，无法进入 `asyncpreempt` 抢占该 goroutine。并且，由于协作式抢占的抢占点检查被 `//go:nosplit` 禁用了，导致协作式和异步抢占都无法抢占该 goroutine。

# 2. 小结

本讲介绍了异步抢占，也就是基于信号的真抢占式调度。至此，我们的 Go runtime 调度器精讲基本结束了，通过十讲内容大致理解了 Go runtime 调度器在做什么。下一讲，会总览全局，把前面讲的内容串起来。

