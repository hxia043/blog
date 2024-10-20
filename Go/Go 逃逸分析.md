### 逃逸分析

逃逸分析是 Go 的特性之一，也是面试中经常容易问到的地方。逃逸分析是编译器为变量确定内存分配的一种分析方式。在编译期间，编译器会确定变量在堆中分配还是栈中分配。

在 C/C++ 中，开发者手动分析变量在内存中的堆还是栈中分配，如果内存在堆中分配，不回收释放的话会影响内存泄漏，这对于开发者来说是非常小心的，也在开发中增添了风险。Go 中将变量分配的任务交给编译器，由编译器在编译期间为变量确定内存的分配。开发者可以更关注于业务，解放了开发者。

同时，编译器分配内存更精确，将不该放到堆中的变量分配到栈中，减轻了垃圾收集器的压力。

### 堆和栈

堆和栈都是内存中的空间。
变量/函数等从高地址到低地址入栈，遵循先进后出的顺序出栈，它是一段固定大小的内存空间，入栈出栈通过 PUSH/POP 指令完成，在指令集层面支持，分配/释放内存非常快。
堆也是内存中的空间，不同的是堆中的内存需要垃圾收集器回收释放，不然将长久存在于堆空间中。如果变量过多的分配到堆空间中也会给垃圾收集器带来压力。

![栈和堆内存分布](./img/内存分配.jpg)

### 逃逸分析场景

编译器遵循三个原则分配变量：
- 如果变量外部是有引用的，则分配到堆区。
- 如果变量是局部变量，无引用，则分配到栈区。
- 如果变量过大，栈去装不下，则分配到堆区。

#### 外部引用

示例如下：
```
func demo() *int {
	a := 1
	return &a
}

func main() {
	b := demo()
	print(b)
}

```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:4:2: a escapes to heap:
./main.go:4:2:   flow: ~r0 = &a:
./main.go:4:2:     from &a (address-of) at ./main.go:5:9
./main.go:4:2:     from return &a (return) at ./main.go:5:2
./main.go:4:2: moved to heap: a
```

可以看到，变量 a 的内存地址被外部引用，编译器将变量 a 分配到堆中。

我们进一步改下代码如下：
```
func demo() int {
	a := 1
	return a
}

func main() {
	b := demo()
	print(b)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go
# 
```

变量 a 传递给变量 b 的是其副本，因此不需要将变量 a 分配到堆中。

闭包也是属于外部引用的一种，举例如下：
```
func demo() func() {
	a := 1

	return func() {
		a++
		print(a)
	}
}

func main() {
	demo()()
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:4:2: demo capturing by ref: a (addr=false assign=true width=8)
./main.go:6:9: func literal escapes to heap:
./main.go:6:9:   flow: ~r0 = &{storage for func literal}:
./main.go:6:9:     from func literal (spill) at ./main.go:6:9
./main.go:6:9:     from return func literal (return) at ./main.go:6:2
./main.go:4:2: a escapes to heap:
./main.go:4:2:   flow: {storage for func literal} = &a:
./main.go:4:2:     from a (captured by a closure) at ./main.go:7:3
./main.go:4:2:     from a (reference) at ./main.go:7:3
./main.go:4:2: moved to heap: a
./main.go:6:9: func literal escapes to heap
```

#### 未知类型

示例：
```
func main() {
	a := 1
	fmt.Println(a)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:7:14: a escapes to heap:
./main.go:7:14:   flow: {storage for ... argument} = &{storage for a}:
./main.go:7:14:     from a (spill) at ./main.go:7:14
./main.go:7:14:     from ... argument (slice-literal-element) at ./main.go:7:13
./main.go:7:14:   flow: {heap} = {storage for ... argument}:
./main.go:7:14:     from ... argument (spill) at ./main.go:7:13
./main.go:7:14:     from fmt.Println(... argument...) (call parameter) at ./main.go:7:13
./main.go:7:13: ... argument does not escape
./main.go:7:14: a escapes to heap
```

变量 a 逃逸到堆中，是因为从编译器角度，fmt.Println() 接收的是接口类型，它并不知道传入的变量是何种类型。

#### 变量申请内存过大

使用 ulimit -a 查看线程可以使用的栈空间：
```
# ulimit -a
stack size                  (kbytes, -s) 8192
```

协程调度到线程上执行，栈空间大小也不会超过操作系统对于线程的限制。

```
func generate8191() {
	nums := make([]int, 8191) // < 64KB
	for i := 0; i < 8191; i++ {
		nums[i] = 1
	}
}

func generate8192() {
	nums := make([]int, 8192) // = 64KB
	for i := 0; i < 8192; i++ {
		nums[i] = 2
	}
}

func generate8193() {
	nums := make([]int, 8193) // = 64KB
	for i := 0; i < 8192; i++ {
		nums[i] = 2
	}
}

func main() {
	generate8191()
	generate8192()
	generate8193()
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:4:14: make([]int, 8191) does not escape
./main.go:11:14: make([]int, 8192) does not escape
./main.go:18:14: make([]int, 8193) escapes to heap:
./main.go:18:14:   flow: {heap} = &{storage for make([]int, 8193)}:
./main.go:18:14:     from make([]int, 8193) (too large for stack) at ./main.go:18:14
./main.go:18:14: make([]int, 8193) escapes to heap
```

可以看到，make([]int, 8193) 超出了栈空间的大小，编译器将其分配到堆中。

#### 大小未知

示例：
```
func main() {
	l := 1
	nums := make([]int, l)
	println(nums)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:5:14: make([]int, l) escapes to heap:
./main.go:5:14:   flow: {heap} = &{storage for make([]int, l)}:
./main.go:5:14:     from make([]int, l) (non-constant size) at ./main.go:5:14
./main.go:5:14: make([]int, l) escapes to heap
```

从编译器角度看，编译器在分配内存时拿到了变量 l，它并不知道 l 是多大，这里的 l 写成 100/1000 对于编译器来说是一样的。因此，对于这种不知道大小的变量分配，编译器将其分配到堆中。

更新下示例：
```
func main() {
	const l = 1
	nums := make([]int, l)
	println(nums)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:5:14: make([]int, l) does not escape
```

#### new

new 创建的变量不一定分配到堆中，这也是编译器在逃逸分析时确定的。示例如下：
```
func main() {
	a := new(int)
	print(a)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:4:10: new(int) does not escape
```

虽然变量是通过 new 创建的，但是变量还是局部变量，编译器将该变量分配到栈中。

#### make

我们在前面的示例中使用 `make` 创建切片，这里我们更新下 `make` 示例如下：
```
func main() {
	m := make([]*int, 10)
	a := 1
	m[0] = &a
	print(a)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:5:2: a escapes to heap:
./main.go:5:2:   flow: {heap} = &a:
./main.go:5:2:     from &a (address-of) at ./main.go:6:9
./main.go:5:2:     from m[0] = &a (assign) at ./main.go:6:7
./main.go:5:2: moved to heap: a
./main.go:4:11: make([]*int, 10) does not escape
```

因为切片中存储的是指向变量的地址。在 Go 中，协程使用的栈空间不足的话会被 Go 运行时拷贝到其它栈空间。这意味着，栈空间的变量是不固定的。  
如果这里将切片中指针指向的变量分配到栈中，则可能会被 Go 的运行时拷贝，导致访问的变量未知。因此，这里变量 a 逃逸到堆中。

换个示例：
```
func main() {
	m := make([]*int, 10)
	print(m)
}
```

输出：
```
# go build -gcflags="-m -l -m" main.go 
# command-line-arguments
./main.go:4:11: make([]*int, 10) does not escape
```

举这个例子是想说明，这里的切片逃逸和切片中指针变量指向的变量逃逸是不一样的。

## 参考文章

- [Go 逃逸分析](https://geektutu.com/post/hpg-escape-analysis.html)
- [go 变量逃逸分析](https://www.cnblogs.com/xingzheanan/p/16082035.html)
- [Go内存分配和逃逸分析-理论篇](https://mp.weixin.qq.com/s?__biz=MzIyNjM0MzQyNg==&mid=2247485774&idx=1&sn=ec7ab97654d105d9cc9661d137ba9b15&scene=21#wechat_redirect)
- [内存逃逸常见问题合集](https://learnku.com/articles/65511)
