# 0. 前言

[上一讲](https://www.cnblogs.com/xingzheanan/p/18407990) 介绍了 Go 程序初始化的过程，这一讲继续往下看，进入调度器的初始化过程。

接着上一讲的执行过程，省略一些不相关的代码，执行到 [runtime/asm_amd64.s:rt0_go:343L](https://github.com/golang/go/blob/master/src/runtime/asm_amd64.s#L343)：
```
(dlv) si
        asm_amd64.s:343 0x45431c*       8b442418        mov eax, dword ptr [rsp+0x18]       // [rsp+0x18] 存储的是 argc 的值，eax = argc
        asm_amd64.s:344 0x454320        890424          mov dword ptr [rsp], eax            // 将 argc 移到 rsp，[rsp] = argc
        asm_amd64.s:345 0x454323        488b442420      mov rax, qword ptr [rsp+0x20]       // [rsp+0x20] 存储的是 argv 的值，rax = [rsp+0x20]
        asm_amd64.s:346 0x454328        4889442408      mov qword ptr [rsp+0x8], rax        // 将 argv 移到 [rsp+0x8]，[rsp+0x8] = argv
        asm_amd64.s:347 0x45432d        e88e2a0000      call $runtime.args                  // 调用 runtime.args 处理栈上的 argc 和 argv
        asm_amd64.s:348 0x454332        e8c9280000      call $runtime.osinit                // 调用 runtime.osinit 初始化系统核心数
        asm_amd64.s:349 0x454337        e8e4290000      call $runtime.schedinit
```

上述指令调用 `runtime.args` 处理函数参数，接着调用 `runtime.osinit` 初始化系统核心数。`runtime.osinit` 在 [runtime.os_linux.go](https://github.com/golang/go/blob/master/src/runtime/os_linux.go#L343) 中定义：
```
func osinit() {
	ncpu = getproccount()
	physHugePageSize = getHugePageSize()
	osArchInit()
}
```

`runtime.osinit` 主要初始化系统核心数 `ncpu`，该核心是逻辑核心数。

接着进入到本文的正题调度器初始化 `runtime.schedinit` 函数。

# 1. 调度器初始化

调度器初始化的代码在 [runtime.schedinit](https://github.com/golang/go/blob/master/src/runtime/proc.go#L793)：  
```
// The bootstrap sequence is:
//
//	call osinit
//	call schedinit
//	make & queue new G
//	call runtime·mstart
//
// The new G calls runtime·main.
func schedinit() {
    // step1: 从 TLS 中获取当前执行线程的 goroutine，gp = m0.tls[0] = g0
    gp := getg()

    // step2: 设置最大线程数
	sched.maxmcount = 10000

    // step3: 初始化线程，这里初始化的是线程 m0
    mcommoninit(gp.m, -1)

    // step4: 调用 procresize 创建 Ps
    procs := ncpu
	if procresize(procs) != nil {
		throw("unknown runnable goroutine during bootstrap")
	}
}
```

省略了函数中不相关的代码。

首先，step1 调用 `getg()` 获取当前线程执行的 goroutine。runtime 中随处可见 `getg()`，它是一个内联的汇编函数，用于直接从当前线程的寄存器或栈 TLS 中获取当前线程执行的 goroutine。Go runtime 会为每个线程（操作系统线程或 Go 运行时线程）维护一个 g 的指针，表示当前线程正在运行的 goroutine。

直观的分析，`get()` 的汇编实现类似于以下内容：
```
TEXT runtime·getg(SB), NOSPLIT, $0
    MOVQ TLS, AX       // 从线程局部存储 (Thread Local Storage) 获取 g
    MOVQ g(AX), BX     // 把 g 的值移动到 BX 寄存器
    RET
```

获取到当前执行 goroutine 之后，在 step3 调用 [mcommoninit](https://github.com/golang/go/blob/master/src/runtime/proc.go#L935) 初始化执行 goroutine 的线程：
```
func mcommoninit(mp *m, id int64) {
    // 获取线程的 goroutine，这里获取的是 g0
	gp := getg()
    ...

    // 对全局变量 sched 加锁
    lock(&sched.lock)

    // 设置 mp 的 id
    if id >= 0 {
		mp.id = id
	} else {
		mp.id = mReserveID()
	}

    // Add to allm so garbage collector doesn't free g->m
	// when it is just in a register or thread-local storage.
	mp.alllink = allm

	// NumCgoCall() iterates over allm w/o schedlock,
	// so we need to publish it safely.
	atomicstorep(unsafe.Pointer(&allm), unsafe.Pointer(mp))         // allm = &m0
	unlock(&sched.lock)
}
```

`mcommoninit` 函数会为 `mp` 设置 id，并且将 mp 和全局变量 allm 关联。更新内存分布如下图：  

![m stack](./img/add%20m%20stack.jpg)

继续执行到 step4 `procresize` 函数，它是 `schedinit` 的重点：
```
func procresize(nprocs int32) *p {
    // old = gomaxprocs = 0
	old := gomaxprocs
	if old < 0 || nprocs <= 0 {
		throw("procresize: invalid arg")
	}

    // procresize 会根据新的 nprocs 调整 P 的数量，这里不做调整，跳过
    if nprocs > int32(len(allp)) {
        ...
    }

    // 初始化 P
	for i := old; i < nprocs; i++ {
		pp := allp[i]
		if pp == nil {
			pp = new(p)
		}
        // 初始化新创建的 P
		pp.init(i)
        // 将新创建的 P 和全局变量 allp 关联
		atomicstorep(unsafe.Pointer(&allp[i]), unsafe.Pointer(pp))      // allp[i] = &pp
	}
    ...
}
```

`procresize` 函数比较长，这里分段介绍。  

首先创建 P，接着调用 `init` 初始化创建的 P：
```
func (pp *p) init(id int32) {
	pp.id = id
	pp.status = _Pgcstop        // _Pgcstop = 3
    ...
}
```

新创建的 P 的 id 是循环的索引 i，状态是 _Pgcstop。接着，将创建的 P 和全局变量 [allp](https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L1165) 进行关联。

接着看 `procresize` 函数：
```
func procresize(nprocs int32) *p {
    // gp = g0
    gp := getg()

    // 判断执行的 goroutine 线程是否绑定到 P 上
    // 如果有，并且是有效的 P，则继续绑定；如果没有，进入 else 逻辑；
	if gp.m.p != 0 && gp.m.p.ptr().id < nprocs {
		// continue to use the current P
		gp.m.p.ptr().status = _Prunning
		gp.m.p.ptr().mcache.prepareForSweep()
    } else {
		...
		gp.m.p = 0              // 初始化 gp.m.p = 0
		pp := allp[0]           // 从 allp 中拿第一个 P
		pp.m = 0                // 设置 P 的 m 等于 0
		pp.status = _Pidle      // 更新 P 的状态为 _Pidle(0)
		acquirep(pp)            // 关联 P 和 m
		...
	}
}
```

`acquirep()` 函数将 P 和当前的线程 m 绑定，如下：
```
func acquirep(pp *p) {
	wirep(pp)
	...
}

func wirep(pp *p) {
    // gp = g0
	gp := getg()

    // 如果当前线程已经绑定了 P 则抛出异常
	if gp.m.p != 0 {
		throw("wirep: already in go")
	}

    // 如果当前 P 已经绑定 m，并且 P 的状态不等于 _Pidle 则抛出异常
	if pp.m != 0 || pp.status != _Pidle {
		id := int64(0)
		if pp.m != 0 {
			id = pp.m.ptr().id
		}
		print("wirep: p->m=", pp.m, "(", id, ") p->status=", pp.status, "\n")
		throw("wirep: invalid p state")
	}

	gp.m.p.set(pp)              // 绑定当前线程 m 的 P 到 pp，这里是 g0.m.p = allp[0]
	pp.m.set(gp.m)              // 绑定 P 的 m 到当前线程，这里是 allp[0].m = m0
	pp.status = _Prunning       // 如果 P 绑定到 m，意味着 P 可以调度 g 在线程上运行了。这里设置 P 的状态为 _Prunning(1)
}
```

根据上述分析，更新内存分布如下图：

![p stack](./img/p%20stack.jpg)  
*(这里我们的 nprocs = 3，所以图中 len(allp) = 3)*

到此还没有结束。继续看 `procresize`：
```
func procresize(nprocs int32) *p {
    ...
    // runnablePs 存储可运行的 Ps
    var runnablePs *p
	for i := nprocs - 1; i >= 0; i-- {
		pp := allp[i]
        // 如果 P 是当前线程绑定的 P 则跳过
		if gp.m.p.ptr() == pp {
			continue
		}
        // 将 P 的状态设为 _Pidle(0)，表示当前 P 是空闲的
		pp.status = _Pidle

        // runqempty 判断 P 中的本地运行队列是否是空队列
        // 如果是空，表明 P 中不存在 goroutine
		if runqempty(pp) {
			pidleput(pp, now)           // 如果是空，将 P 和全局变量 sched 绑定，线程可以通过 sched 找到空闲状态的 P
		} else {
			pp.m.set(mget())            // 如果不为空，调用 mget() 获取空闲的线程 m。并且将 P.m 绑定到该线程
			pp.link.set(runnablePs)     // 将 P 的 link 指向 runnablePs，表明 P 是可运行的
			runnablePs = pp             // 将 runnablePs 指向 P，调用者通过 runnalbePs 拿到可运行的 P
		}
	}

	...
	return runnablePs
}
```

最后的一段就是对 allp 中没有绑定到当前线程的 P 做处理。首先，设置 P 的状态为 _Pidle(0)，接着调用 [runqempty](https://github.com/golang/go/blob/master/src/runtime/proc.go#L6675) 判断当前线程的本地运行队列是否为空：
```
// runqempty reports whether pp has no Gs on its local run queue.
// It never returns true spuriously.
func runqempty(pp *p) bool {
	// Defend against a race where 1) pp has G1 in runqnext but runqhead == runqtail,
	// 2) runqput on pp kicks G1 to the runq, 3) runqget on pp empties runqnext.
	// Simply observing that runqhead == runqtail and then observing that runqnext == nil
	// does not mean the queue is empty.
	for {
		head := atomic.Load(&pp.runqhead)
		tail := atomic.Load(&pp.runqtail)
		runnext := atomic.Loaduintptr((*uintptr)(unsafe.Pointer(&pp.runnext)))
		if tail == atomic.Load(&pp.runqtail) {
			return head == tail && runnext == 0
		}
	}
}
```

这里 P 中的 [runq](https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L644) 存储的是本地运行队列。P 的 runqhead 指向 runq 队列(实际是数组) 的头，runqtail 指向 runq 队尾。  
P 中的 runnext 指向下一个执行的 goroutine，它的优先级是最高的。可以参考 `runqempty` 中的注释去看为什么判断空队列要这么写。

如果 P 中无可运行的 goroutine，则调用 `pidleput` 将 P 添加到全局变量 sched 中：
```
func pidleput(pp *p, now int64) int64 {
	...
	pp.link = sched.pidle           // P.link = shced.pidle             
	sched.pidle.set(pp)             // shced.pidle = P
	sched.npidle.Add(1)             // sched.npidle 表示空间的 P 数量
	...
	return now
}
```

这里我们的 `nprocs = 3`，初始化只有一个 allp[0] 是 _Prunning 的，其余两个 Ps 是 _Pidle 状态。更新内存分布如下图：

![full stack](./img/full%20stack.jpg)

# 2. 小结

好了，到这里我们的调度器初始化逻辑基本介绍完了。下一讲，将继续分析 main gouroutine 的创建。
