# 0. 前言

锁作为并发编程中的关键一环，是应该要深入掌握的。

# 1. 锁

## 1.1 示例

实现锁很简单，示例如下：
```
var global int

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup	

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mu.Lock()
			global++
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println(global)
}
```

输出：
```
2
```

在 goroutine 中给全局变量 `global` 加锁，实现并发顺序增加变量。其中，`sync.Mutex.Lock()` 对变量/临界区加锁，`sync.Mutex.Unlock()` 对变量/临界区解锁。

## 1.2 sync.Mutex

我们看 `sync.Mutex` 互斥锁结构：
```
type Mutex struct {
	state int32
	sema  uint32
}
```

其中，`state` 表示锁的状态，`sema` 表示信号量。

进入 `sync.Mutex.Lock()` 查看加锁的方法。
### 1.2.1 sync.Mutex.Lock()

```
func (m *Mutex) Lock() {
	// Fast path: grab unlocked mutex.
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		return
	}
	// Slow path (outlined so that the fast path can be inlined)
	m.lockSlow()
}
```

首先进入 `Fast path` 逻辑，原子 `CAS` 操作比较锁状态 `m.state` 和 0，如果相等则更新当前锁为已加锁状态。这里锁标志位如下：

![lock state](./img/lock%20state.jpg)

从低（右）到高（左）的三位表示锁状态/唤醒状态/饥饿状态：
```
const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken
	mutexStarving
)
```

标志位初始值为 0，1 表示状态生效。

前三位之后的位数表示排队等待锁的 goroutine 数目，总共可以允许 1<<(32-3) 个 goroutine 等待锁。


这里假设有两个 goroutine G1 和 G2 抢占锁，其中 G1 通过 `Fast path` 获取锁，将锁的状态置为 1。这时候 G2 未获得锁，进入 `Slow path`：
```
func (m *Mutex) lockSlow() {
	var waitStartTime int64
	starving := false
	awoke := false
	iter := 0
	old := m.state
	for {
		// step1: 进入自旋
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
			runtime_doSpin()
			iter++
			old = m.state
			continue
		}

        ...
    }
}
```

`Slow path` 的代码量不大，但涉及状态转换很复杂，不容易看懂。这里拆成每个步骤，根据不同场景分析具体源码。

进入 `Mutex.lockSlow()`，初始化各个状态位，将当前锁状态赋给变量 `old`，进入 for 循环，执行第一步自旋逻辑。自旋会独占 CPU，让 CPU 空跑，但是减少了频繁切换 goroutine 带来的内存/时间消耗。如果使用的适当，会节省 CPU 开销，使用的不适当，会造成 CPU 浪费。这里进入自旋是很严苛的，通过三个条件判断能否自旋：
1. 当前锁是普通模式才能进入自旋。
2. [runtime.sync_runtime_canSpin](https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/proc.go#L6038) 需要返回 true：
   - 当前 goroutine 进入自旋的次数小于 4 次；
   - goroutine 运行在多 CPU 的机器上；
   - 当前机器上至少存在一个正在运行的处理器 P 并且处理的运行队列为空；

假设 G2 可以进入自旋，运行 `runtime_doSpin()`：
```
# src/runtime/lock_futex.go
const active_spin_cnt = 30

# src/runtime/proc.go
//go:linkname sync_runtime_doSpin sync.runtime_doSpin
//go:nosplit
func sync_runtime_doSpin() {
	procyield(active_spin_cnt)
}

# src/runtime/asm_amd64.s
TEXT runtime·procyield(SB),NOSPLIT,$0-0
	MOVL	cycles+0(FP), AX
again:
	PAUSE
	SUBL	$1, AX
	JNZ	again
	RET
```

自旋实际上是 CPU 执行了 30 次 PAUSE 指令。

自旋是在等待，等待锁释放的过程。假设在自旋期间 G1 已释放锁，更新 `m.state` 为 0。那么，在 G2 自旋逻辑中 `old = m.state` 将更新 old 为 0。继续往下看，for 循环中做了什么。
```
func (m *Mutex) lockSlow() {
	...
	for {
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			...
		}

        // step2: 更新 new，这里 new 为 0
        new := old

		// step2: 继续更新 new
        // -      如果锁为普通锁，更新锁状态为已锁。如果锁为饥饿锁，跳过饥饿锁更新。
        // -      这里更新锁为 1
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}

        // step2：继续更新 new
        // -      如果锁为已锁或饥饿的任何一种，则更新 new 的 goroutine 排队等待位
        // -      这里锁为已释放，new 为 1
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}

        // step2: 继续更新 new
        // -      如果 goroutine 处于饥饿状态，并且当前锁是已锁的，更新 new 为饥饿状态
        // -      这里锁为已释放，new 为 1
        if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}

        // step2: 继续更新 new
        // -      如果当前 goroutine 是唤醒的，重置唤醒位为 0
        // -      goroutine 不是唤醒的，new 为 1
        if awoke {
			// The goroutine has been woken from sleep,
			// so we need to reset the flag in either case.
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}

        // step3: CAS 比较 m.state 和 old，如果一致则更新 m.state 到 new
        // -      这里 m.state = 0，old = 0，new = 1
        // -      更新 m.state 为 new，当前 goroutine 获得锁
        if atomic.CompareAndSwapInt32(&m.state, old, new) {
            // 如果更新锁之前的状态不是饥饿或已锁，表示当前 goroutine 已获得锁，跳出循环。
			if old&(mutexLocked|mutexStarving) == 0 {
				break // locked the mutex with CAS
			}
            ...
        }
    }
}
```

这里将自旋后的逻辑简化为两步，更新锁的期望状态 new 和通过原子 CAS 操作更新锁。这里的场景不难，我们可以简化上述流程为如下示意图：  

![lock with spinning](./img/lock%20race%20with%20spinning.jpg)

# 2. 小结

本文介绍了 Go 互斥锁的基本结构，并且给出一个抢占互斥锁的基本场景，通过场景从源码角度分析互斥锁。
