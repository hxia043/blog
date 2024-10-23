sync.WaitGroup 是个比较有意思的包，本来为了面试可以草草过一下，想想这个包短小精悍，而且比较有意思，可以记录下。

sync.WaitGroup 通过维护一个 64 位的 state 和信号量 sema 来实现等待 goroutine 结束。其中 64 位的 goroutine 的前 32 位用来记录 goroutine 的数量，后 32 用来记录等待的 waiter 数量。

sync.WaitGroup 支持 `Add`，`Done` 和 `Wait` 方法。其中 `Add` 方法用来更新 state 的 goroutine 数量。`Done` 方法用来减 state 中 goroutine 的数量。`Wait` 方法是一个阻塞方法，如果未满足 state 中 goroutine 数量为 0，则进入阻塞状态，直到最后一个将 state 中等待的 goroutine 置 0 的 goroutine 在 `Done` 方法中唤醒它。

具体可以参考源码，短小精悍可以多看几遍。

