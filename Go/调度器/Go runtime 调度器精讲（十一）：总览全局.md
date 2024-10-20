# 0. 前言

前面用了十讲介绍了 Go runtime 调度器，这一讲结合一些图在总览下 Go runtime 调度器。

# 1. 状态转换图

首先是 Goroutine 的状态转换图：

![status process](./img/status%20process.jpg)

大部分转移路径前面几讲也介绍过，这里就不继续介绍了(下同)。

接着是 P 的状态转移图：

![P status process](./img/P%20status%20process.jpg)

最后是线程 M 的状态转移图：

![M status process](./img/m%20status%20process.jpg)

（*注：上述图片均来自于公众号 [码农桃花源]，饶大画的太好了，直接让人不想画了。很好的公众号，干货满满，逻辑严密，清晰，偶尔还诙谐幽默，推荐哦。*）

# 2. GPM

## 2.1 GPM 工作流程 

直接上图：

![GPM process](./img/GPM%20process.jpg)

（*这个图来自资深 Gopher 曹大，他的网站在 [这里](https://xargin.com/)，推荐哦*）

## 2.2 调度流程

![schedule process](./img/schedule%20process.jpg)

点这里看 -> [动态演示图](https://www.figma.com/proto/JYM6TcdzBx7WtanhcJX0rP/bootstrap-(Copy)?page-id=5106%3A2&node-id=5106-3&scaling=scale-down)

（*这个图和动画也来自曹大哈哈*）

# 3. 小结

基本到这里 Go runtime 调度器精讲就介绍差不多了，希望这几讲能让大家有所收获，感谢陪伴，再见。

# 4. 参考资料

- [go语言调度器源代码情景分析](https://mp.weixin.qq.com/mp/homepage?__biz=MzU1OTg5NDkzOA==&hid=1&sn=8fc2b63f53559bc0cee292ce629c4788&scene=25#wechat_redirect)
- [The Go scheduler](https://morsmachine.dk/go-scheduler)
- [Go Wiki: Debugging performance issues in Go programs](https://go.dev/wiki/Performance)
- [goroutine 调度器](https://qcrao91.gitbook.io/go/goroutine-tiao-du-qi)
- [Go 语言高级编程](https://www.bookstack.cn/read/advanced-go-programming-book/ch3-asm-readme.md)

