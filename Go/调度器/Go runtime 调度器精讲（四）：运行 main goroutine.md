# 0. 前言

皇天不负有心人，终于我们到了运行 main goroutine 环节了。让我们走起来，看看一个 goroutine 到底是怎么运行的。

# 1. 运行 goroutine

稍微回顾下前面的内容，第一讲 Go 程序初始化，介绍了 Go 程序是怎么进入到 runtime 的，随之揭开 runtime 的面纱。第二讲，介绍了调度器的初始化，要运行 goroutine 调度器是必不可少的，只有调度器准备就绪才能开始工作。第三讲，介绍了 main goroutine 是如何创建出来的，只有创建一个 goroutine 才能开始运行，否则执行代码无从谈起。这一讲，我们继续介绍如何运行 main goroutine。

我们知道 main goroutine 此时处于 _Grunnable 状态，要使得 main goroutine 处于 _Grunning 状态，还需要将它和 P 绑定。毕竟 P 是负责调度任务给线程处理的，只有和 P 绑定线程才能处理相应的 goroutine。

## 1.1 绑定 P

回到代码 `newproc`：
```
func newproc(fn *funcval) {
	gp := getg()
	pc := getcallerpc()
	systemstack(func() {
		newg := newproc1(fn, gp, pc)        // 创建 newg，这里是 main goroutine

		pp := getg().m.p.ptr()              // 获取当前工作线程绑定的 P，这里是 g0.m.p = allp[0]
		runqput(pp, newg, true)             // 绑定 allp[0] 和 main goroutine

		if mainStarted {                    // mainStarted 还未启动，这里是 false
			wakep()
		}
	})
}
```

进入 `runqput` 函数查看 main goroutine 是怎么和 `allp[0]` 绑定的：
```
// runqput tries to put g on the local runnable queue.
// If next is false, runqput adds g to the tail of the runnable queue.
// If next is true, runqput puts g in the pp.runnext slot.
// If the run queue is full, runnext puts g on the global queue.
// Executed only by the owner P.
func runqput(pp *p, gp *g, next bool) {
	...
	if next {
	retryNext:
		oldnext := pp.runnext                                               // 从 P 的 runnext 获取下一个将要执行的 goroutine，这里 pp.runnext = nil
		if !pp.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {         // 将 P 的 runnext 更新为 gp，这里的 gp 是 main goroutine
			goto retryNext  
		}
		if oldnext == 0 {                                                   // 如果 P 原来要执行的 goroutine 是 nil，则直接返回，这里创建的是 main goroutine 将直接返回
			return
		}
		gp = oldnext.ptr()                                                  // 如果不为 nil，表示是一个将要执行的 goroutine。后续对这个被赶走的 goroutine 进行处理
	}

retry:
	h := atomic.LoadAcq(&pp.runqhead)
	t := pp.runqtail
    
	if t-h < uint32(len(pp.runq)) {                                         // P 的队尾和队头指向本地运行队列 runq，如果当前队列长度小于 runq 则将赶走的 goroutine 添加到队尾
		pp.runq[t%uint32(len(pp.runq))].set(gp)
		atomic.StoreRel(&pp.runqtail, t+1)
		return
	}

	if runqputslow(pp, gp, h, t) {                                          // 如果当前 P 的队列长度等于不小于 runq，表示本地队列满了，将赶走的 goroutine 添加到全局队列中
		return
	}

	goto retry
}
```

`runqput` 函数绑定 P 和 goroutine，同时处理 P 中的本地运行队列。基本流程在注释中已经介绍的比较清楚了。

这里我们绑定的是 main goroutine，直接绑定到 P 的 runnext 成员即可。不过对于 `runqput` 的整体处理来说，还需要在介绍一下 `runqputslow` 函数：
```
// Put g and a batch of work from local runnable queue on global queue.
// Executed only by the owner P.
func runqputslow(pp *p, gp *g, h, t uint32) bool {
	var batch [len(pp.runq)/2 + 1]*g                                                // 定义 batch，长度是 P.runq 的一半。batch 用来装 g

	// First, grab a batch from local queue.
	n := t - h
	n = n / 2
	if n != uint32(len(pp.runq)/2) {
		throw("runqputslow: queue is not full")
	}
	for i := uint32(0); i < n; i++ {
		batch[i] = pp.runq[(h+i)%uint32(len(pp.runq))].ptr()                        // 从 P 的 runq 中拿出一半的 g 到 batch 中
	}
	if !atomic.CasRel(&pp.runqhead, h, h+n) { // cas-release, commits consume       // 更新 P 的 runqhead 的指向，它指向的是本地队列的头
		return false
	}
	batch[n] = gp                                                                   // 将赶走的 goroutine 放到 batch 尾

	if randomizeScheduler {                                                         // 如果是随机调度的话，这里还要打乱 batch 中 g 的顺序以保证随机性
		for i := uint32(1); i <= n; i++ {
			j := fastrandn(i + 1)
			batch[i], batch[j] = batch[j], batch[i]
		}
	}

	// Link the goroutines.
	for i := uint32(0); i < n; i++ {
		batch[i].schedlink.set(batch[i+1])                                          // batch 中 goroutine 的 schedlink 按顺序指向其它 goroutine，构造一个链表
	}
	var q gQueue                                                                    // gQueue 是一个包含头和尾的指针，将头和尾指针分别指向 batch 的头 batch[0] 和尾 batch[n]
	q.head.set(batch[0])
	q.tail.set(batch[n])

	// Now put the batch on global queue.
	lock(&sched.lock)                                                               // 操作全局变量 sched，为 sched 加锁
	globrunqputbatch(&q, int32(n+1))                                                // globrunqputbatch 将 q 指向的 batch 传给全局变量 sched
	unlock(&sched.lock)                                                             // 解锁
	return true
}

func globrunqputbatch(batch *gQueue, n int32) {
	assertLockHeld(&sched.lock)

	sched.runq.pushBackAll(*batch)                                                  // 这里将 sched.runq 指向 batch
	sched.runqsize += n                                                             // sched 的 runqsize 加 n，n 表示新添加进 sched.runq 的 goroutine
	*batch = gQueue{}
}
```

如果 P 的本地队列已满，则在 `runqputslow` 中拿出本地队列的一半 goroutine 放到 sched.runq 全局队列中。这里本地队列是固定长度，容量有限，用数组来表示队列。而全局队列长度是不固定的，用链表来表示全局队列。

我们可以画出示意图如下图，注意示意图只是加深理解，和我们这里运行 main goroutine 的流程没关系：

![sched runq](./img/sched%20runq.jpg)

## 1.2 运行 main goroutine

P 和 main goroutine 绑定之后，理论上已经可以运行 main goroutine 了。继续看代码执行的什么：
```
> runtime.rt0_go() /usr/local/go/src/runtime/asm_amd64.s:358 (PC: 0x45434a)
Warning: debugging optimized function
   353:         PUSHQ   AX
   354:         CALL    runtime·newproc(SB)
   355:         POPQ    AX
   356:
   357:         // start this M
=> 358:         CALL    runtime·mstart(SB)      // 调用 mstart 意味着当前线程开始工作了；mstart 是一个永不返回的函数
   359:
   360:         CALL    runtime·abort(SB)       // mstart should never return
   361:         RET
   362:
```

向下执行：
```
(dlv) si
> runtime.mstart() /usr/local/go/src/runtime/asm_amd64.s:394 (PC: 0x4543c0)
Warning: debugging optimized function
TEXT runtime.mstart(SB) /usr/local/go/src/runtime/asm_amd64.s
=>      asm_amd64.s:394 0x4543c0        e87b290000      call $runtime.mstart0
        asm_amd64.s:395 0x4543c5        c3              ret
```

调用 `runtime.mstart0`：
```
func mstart0() {
	gp := getg()                // gp = g0
    ...
    mstart1()
    ...
}
```

调用 `mstart1`：
```
func mstart1() {
	gp := getg()                                    // gp = g0

    // 保存线程执行的栈，当线程进入 schedule 函数就不会返回，这意味着线程执行的栈是可复用的
    gp.sched.g = guintptr(unsafe.Pointer(gp))
	gp.sched.pc = getcallerpc()
	gp.sched.sp = getcallersp()

    ...
    if fn := gp.m.mstartfn; fn != nil {             // 执行 main goroutine，fn == nil
		fn()
	}

    ...
    schedule()                                      // 线程进入 schedule 调度循环，该循环是永不返回的
}
```

进入 `schedule`：
```
func schedule() {
	mp := getg().m                                  // mp = m0
    ...
top:
	pp := mp.p.ptr()                                // pp = allp[0]
	pp.preempt = false

    // 线程有两种状态，自旋和非自旋。自旋表示线程没有工作，在找工作阶段。非自旋表示线程正在工作
    // 这里如果线程自旋，但是线程绑定的 P 本地队列有 goroutine 则报异常
    if mp.spinning && (pp.runnext != 0 || pp.runqhead != pp.runqtail) {
		throw("schedule: spinning with local work")
	}

    // blocks until work is available
    gp, inheritTime, tryWakeP := findRunnable()     // 找一个处于 _Grunnable 状态的 goroutine 出来

    ...
    execute(gp, inheritTime)                        // 运行该 goroutine，这里运行的是 main goroutine
}
```

`schedule` 中的重点是 `findRunaable` 函数，进入该函数：
```
// Finds a runnable goroutine to execute.
// Tries to steal from other P's, get g from local or global queue, poll network.
// tryWakeP indicates that the returned goroutine is not normal (GC worker, trace
// reader) so the caller should try to wake a P.
func findRunnable() (gp *g, inheritTime, tryWakeP bool) {
	mp := getg().m                      // mp = m0

top:
	pp := mp.p.ptr()                    // pp = allp[0] = p0

    ...
    // Check the global runnable queue once in a while to ensure fairness.
    // Otherwise two goroutines can completely occupy the local runqueue
    // by constantly respawning each other.
    // 官方的注释对这一段逻辑已经解释的很详细了，我们就跳过了，偷个懒
    if pp.schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)
		gp := globrunqget(pp, 1)
		unlock(&sched.lock)
		if gp != nil {
			return gp, false, false
		}
    }

    // local runq
    // 从 P 的本地队列找 goroutine
	if gp, inheritTime := runqget(pp); gp != nil {
		return gp, inheritTime, false
	}
    ...
}
```

`findRunnable` 中首先为了公平，每调用 schedule 函数 61 次就要从全局可运行队列中获取 goroutine，防止全局队列中的 goroutine 被“饿死”。接着从 P 的本地队列中获取 goroutine，这里运行的是 main goroutine 将从 P 的本地队列中获取 goroutine。查看 `runqget`：
```
func runqget(pp *p) (gp *g, inheritTime bool) {
	// If there's a runnext, it's the next G to run.
	next := pp.runnext
	// If the runnext is non-0 and the CAS fails, it could only have been stolen by another P,
	// because other Ps can race to set runnext to 0, but only the current P can set it to non-0.
	// Hence, there's no need to retry this CAS if it fails.
	if next != 0 && pp.runnext.cas(next, 0) {
		return next.ptr(), true
	}

	for {
		h := atomic.LoadAcq(&pp.runqhead) // load-acquire, synchronize with other consumers
		t := pp.runqtail
		if t == h {
			return nil, false
		}
		gp := pp.runq[h%uint32(len(pp.runq))].ptr()
		if atomic.CasRel(&pp.runqhead, h, h+1) { // cas-release, commits consume
			return gp, false
		}
	}
}
```

注释已经比较详细了，首先拿到 P 的 runnext 作为要运行的 goroutine。如果拿到的 goroutine 不是空，则重置 runnext，并且返回拿到的 goroutine。如果拿到的 goroutine 是空的，则从本地队列中拿 goroutine。

通过 `findRunnable` 我们拿到可执行的 main goroutine。接着调用 `execute` 执行 main goroutine。


进入 `execute`：
```
func execute(gp *g, inheritTime bool) {
	mp := getg().m                                  // mp = m0

    mp.curg = gp                                    // mp.curg = g1
	gp.m = mp                                       // gp.m = m0
	casgstatus(gp, _Grunnable, _Grunning)           // 更新 goroutine 的状态为 _Grunning
	gp.waitsince = 0
	gp.preempt = false
	gp.stackguard0 = gp.stack.lo + stackGuard
	if !inheritTime {
		mp.p.ptr().schedtick++
	}

    ...
    gogo(&gp.sched)                             
}
```

在 `execute` 中将线程和 gouroutine 关联起来，更新 goroutine 的状态，然后调用 gogo 完成从 g0 栈到 gp 栈的切换，gogo 是用汇编编写的，原因如下：
```
gogo 函数也是通过汇编语言编写的，这里之所以需要使用汇编，是因为 goroutine 的调度涉及不同执行流之间的切换。

前面我们在讨论操作系统切换线程时已经看到过，执行流的切换从本质上来说就是 CPU 寄存器以及函数调用栈的切换，然而不管是 go 还是 c 这种高级语言都无法精确控制 CPU 寄存器，因而高级语言在这里也就无能为力了，只能依靠汇编指令来达成目的。
```

进入 `gogo`，`gogo` 传入的是 goroutine 的 sched 结构：
```
TEXT runtime·gogo(SB), NOSPLIT, $0-8
	MOVQ	buf+0(FP), BX		                // gobuf
	MOVQ	gobuf_g(BX), DX                     // gobuf 的 g 赋给 DX
	MOVQ	0(DX), CX		                    // make sure g != nil
	JMP	gogo<>(SB)                              // 跳转到私有函数 gogo<>

TEXT gogo<>(SB), NOSPLIT, $0
	get_tls(CX)                                 // 获取当前线程 tls 中的 goroutine
	MOVQ	DX, g(CX)
	MOVQ	DX, R14		                        // set the g register
	MOVQ	gobuf_sp(BX), SP	                // restore SP
	MOVQ	gobuf_ret(BX), AX                   // AX = gobuf.ret
	MOVQ	gobuf_ctxt(BX), DX                  // DX = gobuf.ctxt
	MOVQ	gobuf_bp(BX), BP                    // BP = gobuf.bp
	MOVQ	$0, gobuf_sp(BX)	                // clear to help garbage collector
	MOVQ	$0, gobuf_ret(BX)
	MOVQ	$0, gobuf_ctxt(BX)
	MOVQ	$0, gobuf_bp(BX)
	MOVQ	gobuf_pc(BX), BX                    // BX = gobuf.pc
	JMP	BX                                      // 跳转到 gobuf.pc 
```

在 `gogo<>` 中完成 g0 到 gp 栈的切换：`MOVQ	gobuf_sp(BX), SP`，并且跳转到 gobuf.pc 执行。我们看 gobuf.pc 要执行的指令地址是什么：
```
asm_amd64.s:421 0x45363a        488b5b08                mov rbx, qword ptr [rbx+0x8]
=>      asm_amd64.s:422 0x45363e        ffe3                    jmp rbx
(dlv) regs
    Rbx = 0x000000000042ee80
```

执行 `JMP BX` 跳转到 `0x000000000042ee80`：
```
(dlv) si
> runtime.main() /usr/local/go/src/runtime/proc.go:144 (PC: 0x42ee80)
Warning: debugging optimized function
TEXT runtime.main(SB) /usr/local/go/src/runtime/proc.go
=>      proc.go:144     0x42ee80        4c8d6424e8      lea r12, ptr [rsp-0x18]
```

终于我们揭开了它的神秘面纱，这个指令指向的是 `runtime.main` 函数的第一条汇编指令。也就是说，跳转到了 `runtime.main`，这个函数会调用我们 main 包下的 main 函数。查看 `runtime.main` 函数：
```
// The main goroutine.
func main() {
	mp := getg().m                          // mp = m0

    if goarch.PtrSize == 8 {
		maxstacksize = 1000000000           // 扩栈，栈的最大空间是 1GB
	} else {
		maxstacksize = 250000000
	}

    ...
    // Allow newproc to start new Ms.
	mainStarted = true

    if GOARCH != "wasm" { // no threads on wasm yet, so no sysmon
		systemstack(func() {
			newm(sysmon, nil, -1)           // 开启监控线程，这个线程很重要，我们后续会讲，这里先放着，让 sysmon 飞一会儿
		})
	}

    ...
    // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
    fn := main_main                         // 这里的 main_main 链接的是 main 包中的 main 函数
	fn()                                    // 执行 main.main
    ...
    runExitHooks(0)

	exit(0)                                 // 执行完 main.main 之后调用 exit 退出线程
    for {
		var x *int32
		*x = 0
	}
}
```

`runtime.main` 是在 main goroutine 栈中执行的。在函数中调用 main.main 执行我们写的用户代码：
```
(dlv) n
266:            fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
=> 267:         fn()
(dlv) s
> main.main() ./hello.go:3 (PC: 0x45766a)
Warning: debugging optimized function
     1: package main
     2:
=>   3: func main() {
     4:         println("Hello World")
     5: }
```

`main.main` 执行完之后线程调用 `exit(0)` 退出程序。

# 2. 小结

至此我们的 main goroutine 就执行完了，花了四讲才算走通了一个 main goroutine，真不容易呀。当然，关于 Go runtime 调度器的故事还没结束，下一讲我们继续。
