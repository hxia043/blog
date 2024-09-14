# 0. 前言

在 [Go runtime 调度器精讲（三）：main goroutine 创建](https://www.cnblogs.com/xingzheanan/p/18411879) 介绍了 main goroutine 的创建，文中我们说 main goroutine 和非 main goroutine 有区别。当时卖了个关子并未往下讲，这一讲我们会继续介绍非 main goroutine (也就是 go 关键字创建的 goroutine，后文统称为 gp) 的运行，并且把这个关子解开，说一说它们的区别在哪儿。

# 1. gp 的创建

首先看一个示例：
```
func g2() {
    time.Sleep(10 * time.Second)
	println("hello world")
}

func main() {
	go g2()

	time.Sleep(1 * time.Minute)
	println("main exit")
}
```

main 函数创建两个 goroutine，一个 main goroutine，一个普通 goroutine。从 [Go runtime 调度器精讲（四）：运行 main goroutine](https://www.cnblogs.com/xingzheanan/p/18412514) 可知 main goroutine 运行完之后就调用 `exit(0)` 退出了。为了能进入 gp，我们这里在 main goroutine 中加了 1 分钟的等待时间。

Go runtime 的启动在前几讲都有介绍，这里直接进入 main 函数，查看 gp 是如何创建的：
```
(dlv) c
> main.main() ./goexit.go:12 (hits goroutine(1):1 total:1) (PC: 0x46238a)
     7: func g2() {
     8:         time.Sleep(10 * time.Second)
     9:         println("hello world")
    10: }
    11:
=>  12: func main() {
    13:         go g2()
    14:
    15:         time.Sleep(30 * time.Minute)
    16:         println("main exit")
    17: }
```

直接看 main 函数，我们看不出 `go` 关键字做了什么，查看 CPU 的汇编指令：
```
(dlv) si
> main.main() ./goexit.go:13 (PC: 0x462395)
        goexit.go:12    0x462384        7645                    jbe 0x4623cb
        goexit.go:12    0x462386        55                      push rbp
        goexit.go:12    0x462387        4889e5                  mov rbp, rsp
        goexit.go:12    0x46238a*       4883ec10                sub rsp, 0x10
        goexit.go:13    0x46238e        488d050b7a0100          lea rax, ptr [rip+0x17a0b]
=>      goexit.go:13    0x462395        e8c6b1fdff              call $runtime.newproc
        goexit.go:15    0x46239a        48b800505c18a3010000    mov rax, 0x1a3185c5000
        goexit.go:15    0x4623a4        e8b79fffff              call $time.Sleep
```

可以看到，`go` 关键字被编译转换后实际调用的是 `$runtime.newproc` 函数，这个函数在 [Go runtime 调度器精讲（四）：运行 main goroutine](https://www.cnblogs.com/xingzheanan/p/18412514) 已经非常详细的介绍过了，这里就不赘述了。

有必要在说明的是，main goroutine 和普通 goroutine 执行的顺序。当调用 `runtime.newproc` 后，gp 被添加到 P 的可运行队列（如果队列满，被添加到全局队列），接着线程会调度运行该 gp。不过对于 `newproc` 来说，gp 放入队列后，`newproc` 就退出了。接着执行后续的 main goroutine 代码。

如果此时 gp 未运行或者未结束，并且 main goroutine 未等待/阻塞的话，main goroutine 将直接退出。

# 2. gp 的退出

前面说 gp 和 main goroutine 的区别主要体现在 goroutine 的退出这里。main goroutine 的退出比较残暴，直接调用 `exit(0)` 退出进程。那么，gp 是怎么退出的呢？

我们在 `g2` 结束点处打断点，看看 `g2` 是怎么退出的：
```
(dlv) b ./goexit.go:10
Breakpoint 1 set at 0x46235b for main.g2() ./goexit.go:10
(dlv) c
hello world
> main.g2() ./goexit.go:10 (hits goroutine(5):1 total:1) (PC: 0x46235b)
     7: func g2() {
     8:         time.Sleep(10 * time.Second)
     9:         println("hello world")
=>  10: }
    11:
    12: func main() {
    13:         go g2()
    14:
    15:         time.Sleep(30 * time.Minute)
(dlv) si
> main.g2() ./goexit.go:10 (PC: 0x46235f)
        goexit.go:9     0x462345        488d05b81b0100  lea rax, ptr [rip+0x11bb8]
        goexit.go:9     0x46234c        bb0c000000      mov ebx, 0xc
        goexit.go:9     0x462351        e88a30fdff      call $runtime.printstring
        goexit.go:9     0x462356        e86528fdff      call $runtime.printunlock
        goexit.go:10    0x46235b*       4883c410        add rsp, 0x10
=>      goexit.go:10    0x46235f        5d              pop rbp
        goexit.go:10    0x462360        c3              ret
        goexit.go:7     0x462361        e89ab1ffff      call $runtime.morestack_noctxt
        goexit.go:7     0x462366        ebb8            jmp $main.g2
```

CPU 执行指令到 `pop rbp`，接着执行 ret:
```
        goexit.go:10    0x46235f        5d              pop rbp
=>      goexit.go:10    0x462360        c3              ret
        goexit.go:7     0x462361        e89ab1ffff      call $runtime.morestack_noctxt
        goexit.go:7     0x462366        ebb8            jmp $main.g2
(dlv) si
> runtime.goexit() /usr/local/go/src/runtime/asm_amd64.s:1651 (PC: 0x45d7a1)
Warning: debugging optimized function
TEXT runtime.goexit(SB) /usr/local/go/src/runtime/asm_amd64.s
        asm_amd64.s:1650        0x45d7a0        90              nop
=>      asm_amd64.s:1651        0x45d7a1        e8ba250000      call $runtime.goexit1
        asm_amd64.s:1653        0x45d7a6        90              nop
```

我们看到了什么，执行 ret 直接跳转到了 `call $runtime.goexit1`。还记得在 [Go runtime 调度器精讲（三）：main goroutine 创建](https://www.cnblogs.com/xingzheanan/p/18411879) 中说每个 goroutine 栈都会在“栈顶”放 `funcPC(goexit) + 1` 的地址。这里实际是做了一个偷梁换柱，gp 的栈在退出执行 ret 时都会跳转到 `call $runtime.goexit1` 继续执行。

进入 `runtime.goexit1`：
```
// Finishes execution of the current goroutine.
func goexit1() {
	...
	mcall(goexit0)                          // mcall 会切换当前栈到 g0 栈，接着在 g0 栈执行 goexit0
}
```

实际执行的是 `goexit0`：
```
// goexit continuation on g0.
func goexit0(gp *g) {
    mp := getg().m                          // 这里是 g0 栈，mp = m0
	pp := mp.p.ptr()                        // m0 绑定的 P

    casgstatus(gp, _Grunning, _Gdead)       // 将 gp 的状态更新为 _Gdead
    gp.m = nil                              // 将 gp 绑定的线程更新为 nil，和线程解绑
    ...

    dropg()                                 // 将当前线程和 gp 解绑
    ...
    gfput(pp, gp)                           // 退出的 gp 还是可以重用的，gfput 将 gp 放到本地或者全局空闲队列中

    ...
    schedule()                              // 线程执行完一个 gp 还没有退出，继续进入 schedule 找 goroutine 执行
}
```

gp 退出了，线程并没有退出，线程将 gp 安顿好之后，继续开始新一轮调度，真是劳模啊。

# 3. 小结

本讲介绍了用 `go` 关键字创建的 goroutine 是如何运行的，下一讲我们放松放松，看几个案例分析调度器的行为。

