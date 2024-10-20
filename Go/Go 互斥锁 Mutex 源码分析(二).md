# 0. 前言

在 [Go 互斥锁 Mutex 源码分析(一)](https://www.cnblogs.com/xingzheanan/p/18376083) 一文中分析了互斥锁的结构和基本的抢占互斥锁的场景。在学习锁的过程中，看的不少文章是基于锁的状态解释的，个人经验来看，从锁的状态出发容易陷入细节，了解锁的状态转换过一段时间就忘，难以做到真正的理解。想来是用静态的方法分析动态的问题导致的。在实践中发现结合场景分析互斥锁对笔者来说更加清晰，因此有了 [Go 互斥锁 Mutex 源码分析(一)](https://www.cnblogs.com/xingzheanan/p/18376083)，本文接着结合不同场景分析互斥锁。

# 1. 不同场景下的锁状态

## 1.1 唤醒 goroutine

给出示意图：  

![lock with awoke](./img/lock%20race%20with%20awoke.jpg)

G1 通过 `Fast path` 拿到锁，G2 在自旋之后，锁还是已锁状态。这是和 [Go 互斥锁 Mutex 源码分析(一)](https://www.cnblogs.com/xingzheanan/p/18376083) 中的场景不一样的地方。接着自旋之后看，这种场景下会发生什么：
```
func (m *Mutex) lockSlow() {
	...
	for {
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
            ...
        }
        // step2: 当前锁未释放，old = 1
        new := old

        // step2: 如果当前锁是饥饿的，跳过期望状态 new 的更新
        // -      这里锁不是饥饿锁，new = old = 1
        if old&mutexStarving == 0 {
			new |= mutexLocked
		}

        // step2: 当前锁未释放，更新 new
        // -      更新 new 的等待 goroutine 位，表示有一个 goroutine 等待
        // -      更新 new 为 1001，new = 9 
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}

		// step2: 当前 goroutine 不是饥饿状态，跳过 new 更新
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}

        // step2: 当前 goroutine 不是唤醒状态，跳过 new 更新
        if awoke {
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}

        // step3: 原子 CAS 更新锁的状态
        // -      这里更新锁 m.state = 1 为 m.state = new = 9
        // -      表示当前有一个 goroutine 在等待锁
        if atomic.CompareAndSwapInt32(&m.state, old, new) {
            ...
            // waitStartTime = 0, queueLifo = false
            queueLifo := waitStartTime != 0
			if waitStartTime == 0 {
                // 更新 waitStartTime
				waitStartTime = runtime_nanotime()
			}

            // step4: 调用 runtime_SemacquireMutex 阻塞 goroutine
            runtime_SemacquireMutex(&m.sema, queueLifo, 1)
            starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
            ...
        }
    }
}
```

`Mutex.lockSlow` 中更新了锁状态，接着进入 `runtime_SemacquireMutex`。`runtime_SemacquireMutex` 是个非常重要的函数，我们有必要介绍它。

`runtime_SemacquireMutex` 接收三个参数。其中，重点是信号量 `&m.sema` 和 `queueLifo`。如果 `queueLifo = false`，当前 goroutine 将被添加到等待锁队列的队尾，阻塞等待唤醒。

G2 执行到 `runtime_SemacquireMutex` 时将进入阻塞等待唤醒状态，那么怎么唤醒 G2 呢？ 我们需要看解锁过程。

### 1.1.1 sync.Mutex.Unlock

在 G2 阻塞等待唤醒时，G1 开始释放锁。进入 `sync.Mutex.Unlock`：
```
func (m *Mutex) Unlock() {
	...
	// 将 m.state 的锁标志位置为 0，表示锁已释放
	new := atomic.AddInt32(&m.state, -mutexLocked)
    // 检查 new 是否为 0，如果为 0 则表示当前无 goroutine 等待，直接退出
    // 这里 new = 9，G2 在等待唤醒
	if new != 0 {
		m.unlockSlow(new)
	}
}
```

进入 `Mutex.unlockSlow`：
```
func (m *Mutex) unlockSlow(new int32) {
    // 检查锁是否已释放，释放一个已经释放的锁将报错
	if (new+mutexLocked)&mutexLocked == 0 {
		fatal("sync: unlock of unlocked mutex")
	}

    // 检查锁是普通锁还是饥饿锁
    if new&mutexStarving == 0 {
        // 这里 new = 8 是普通锁，进入处理普通锁逻辑
		old := new
		for {
            // 如果没有 goroutine 等待，则返回
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return
			}

            // old 的唤醒位置 1，并且将等待的 goroutine 减 1，表示将唤醒一个等待中的 goroutine
            // 这里 new = 2
			new = (old - 1<<mutexWaiterShift) | mutexWoken
            // m.state = 8, old = 8, new = 2
            // CAS 更新 m.state = new = 2
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
                // 进入 runtime_Semrelease 唤醒 goroutine
				runtime_Semrelease(&m.sema, false, 1)
				return
			}
			old = m.state
		}
	} else {
        // 处理饥饿锁逻辑，暂略
		runtime_Semrelease(&m.sema, true, 1)
	}
}
```

`sync.Mutex.Unlock` 中的 `runtime_Semrelease` 唤醒队列中等待的 goroutine。其中，主要接收信号量 `&m.sema` 和 `handoff` 两个参数。这里 `handoff = false`，将增加信号量，唤醒队列中等待的 goroutine G2。

### 1.1.2 唤醒 G2

唤醒之后，G2 继续执行后续代码：
```
func (m *Mutex) lockSlow() {
	...
	for {
		...
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			...
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)

			// 检查唤醒的 goroutine 是否是饥饿模式
			// 如果是饥饿模式，或等待锁时间超过 1ms 则将 goroutine 置为饥饿模式
			// 注意这是 goroutine 是饥饿的，不是锁是饥饿锁
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
			
			// m.state 在 G1 unlock 时被更新为 2
			old = m.state

			// 锁不是饥饿锁，跳过
			if old&mutexStarving != 0 {
				...
			}
			awoke = true
			iter = 0
		}
	}
}
```

唤醒后的 G2 将 old 更新为 2。信号量增加，释放锁，只会唤醒一个 goroutine，被唤醒的 goroutine，这里是 G2，将继续循环：
```
func (m *Mutex) lockSlow() {
	...
	for {
		// old = 2，不会进入自旋
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			...
		}
		// 更新 new：new 是期望 goroutine 更新的状态
		// 这里 new = old = 2
		new := old

		// old = 2，不是饥饿锁
		// 更新 new 为 011，3
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}
		// old = 2，表示锁已释放，不会将 goroutine 加入等待位
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}
		// 不饥饿，跳过
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}
		// awoke = true
		if awoke {
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			// 重置唤醒位，将 new 更新为 001，1
			new &^= mutexWoken
		}

		// m.state = 2, old = 2, new =1
		// CAS 更新 m.state= new = 1，表示当前 goroutine 已加锁
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			// 当前 goroutine 已加锁跳出循环
			if old&(mutexLocked|mutexStarving) == 0 {
				break // locked the mutex with CAS
			}
			...
		}
	}
}
```

在循环一轮后，G2 将拿到锁，接着执行临界区代码，最后在释放锁。


这里的场景是唤醒之后，goroutine 不饥饿。那么饥饿锁又是如何触发的呢？我们继续看饥饿锁的场景。

## 1.2 饥饿锁

饥饿锁场景下的示意图如下：  

![lock with starving](./img/lock%20with%20starving.jpg)

当 G1 释放锁时，G3 正在自旋等待锁释放。当 G1 释放锁时，被唤醒的 G2 和自旋的 G3 竞争大概率会拿不到锁。Go 在 1.9 中引入互斥锁的 [饥饿模式](https://github.com/golang/go/commit/0556e26273f704db73df9e7c4c3d2e8434dec7be) 来确保互斥锁的公平性。

对于互斥锁循环中的大部分流程，我们在前两个场景下也过了一遍，这里有重点的摘写，以防赘述。

首先，还是看 G2，当 G1 释放锁时，G2 被唤醒，执行后续代码。如下：
```
func (m *Mutex) lockSlow() {
	...
	for {
		...
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			...
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)

			// 唤醒 G2，G2 等待锁时间超过 1ms
			// starving = true
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs

			// 锁被 G3 抢占，m.state = 0011
			old = m.state

			// 这时候 old 还不是饥饿锁，跳过
			if old&mutexStarving != 0 {
				...
			}
			awoke = true
			iter = 0
		}
	}
}
```

唤醒 G2 之后，G2 等待锁时间超过 1ms 进入饥饿模式。接着进入下一轮循环：
```
func (m *Mutex) lockSlow() {
	...
	for {
		// old 是唤醒锁，不会进入自旋
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			...
		}

		// 锁的期望状态，new = old = 0011
		new := old

		// 锁不是饥饿锁，更新 new 的锁标志位为已锁
		// new = 0011
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}

		// 锁如果是饥饿或者已锁状态更新 goroutine 等待位
		// new = 1011
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}

		// goroutine 饥饿，且锁已锁
		// 更新 new 为饥饿状态，new = 1111
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}

		// 这里 G2 是唤醒的，重置唤醒位
		// new = 1101
		if awoke {
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}

		// CAS 更新 m.state = new = 1101
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			...
			// G2 入队列过，这里 queueLifo = true
			queueLifo := waitStartTime != 0

			// 将 G2 重新加入队列，并加入到队首，阻塞等待
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)
			...
		}
	}
}
```

G2 进入饥饿模式，将互斥锁置为饥饿模式，当前互斥锁状态为 m.state = 1101。G2 作为队列中的队头，阻塞等待锁释放。

类似的，我们看 G3 释放锁的过程。

### 1.2.1 释放饥饿锁

G3 开始释放锁：
```
func (m *Mutex) Unlock() {
	...

	// new = 1100
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		// 进入 Mutex.unlockSlow
		m.unlockSlow(new)
	}
}

func (m *Mutex) unlockSlow(new int32) {
	...
	// new = 1100，是饥饿锁
	if new&mutexStarving == 0 {
		...
	} else {
		// 进入处理饥饿锁逻辑
		// handoff = true，直接将队头阻塞的 goroutine 唤醒
		runtime_Semrelease(&m.sema, true, 1)
	}
}
```

### 1.2.2 饥饿锁唤醒

在一次的在队头中阻塞的 G2 被唤醒，接着执行唤醒后的代码：
```
func (m *Mutex) lockSlow() {
	...
	for {
		...
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			...
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
			old = m.state

			// old = 1100，是饥饿锁
			if old&mutexStarving != 0 {
				...

				// delta = -(1001)
				delta := int32(mutexLocked - 1<<mutexWaiterShift)
				if !starving || old>>mutexWaiterShift == 1 {
					...
					// delta = -(1101)
					delta -= mutexStarving
				}

				//更新互斥锁状态 m.state = 0001，退出循环
				atomic.AddInt32(&m.state, delta)
				break
			}
		}
	}
}
```

唤醒之后的 G2 直接获得锁，将互斥锁状态置为已锁，直到释放。

# 2. 锁状态流程

前面我们根据几个场景给出了互斥锁的状态转换过程，这里直接给出互斥锁的流程图如下：

![lock with procedure](./img/lock%20with%20procedure.jpg)

# 3. 总结

本文是 Go 互斥锁 Mutex 源码分析的第二篇，进一步通过两个场景分析互斥锁的状态转换。互斥锁的状态转换如果陷入状态更新，很容易头晕，这里通过不同场景，逐步分析，整个状态，接着给出状态转换流程图，力图做到源码层面了解锁的状态转换。

