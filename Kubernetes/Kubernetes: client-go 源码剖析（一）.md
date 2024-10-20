# 0. 前言

在看 `kube-scheduler` 组件的过程中遇到了 `kube-scheduler` 对于 `client-go` 的调用，泛泛的理解调用过程总有种隔靴搔痒的感觉，于是调转头先把 `client-go` 理清楚在回来看 `kube-scheduler`。

为什么要看 `client-go`，并且要深入到原理，源码层面去看。很简单，因为它很重要。重要在两方面：
1. `kubernetes` 组件通过 `client-go` 和 `kube-apiserver` 交互。
2. `client-go` 简单，易用，大部分基于 `Kubernetes` 做二次开发的应用，在和 `kube-apiserver` 交互时会使用 `client-go`。

当然，不仅在于使用，理解层面，对于我们学习代码开发，架构等也有帮助。

# 1. client-go 客户端对象

`client-go` 支持四种客户端对象，分别是 `RESTClient`，`ClientSet`，`DynamicClient` 和 `DiscoveryClient`：

![client-go 客户端](img/client-go%20客户端.png)

组件或者二次开发的应用可以通过这四种客户端对象和 `kube-apiserver` 交互。其中，`RESTClient` 是最基础的客户端对象，它封装了 `HTTP Request`，实现了 `RESTful` 风格的 `API`。`ClientSet` 基于 `RESTClient`，封装了对于 `Resource` 和 `Version` 的请求方法。`DynamicClient` 相比于 `ClientSet` 提供了全资源，包括自定义资源的请求方法。`DiscoveryClient` 用于发现 `kube-apiserver` 支持的资源组，资源版本和资源信息。

每种客户端适用的场景不同，主要是对 `HTTP Request` 做了层层封装，具体的代码实现可参考 [client-go 客户端](https://github.com/hxia043/go-by-example/tree/main/kubernetes/client-go)。

# 2. informer 机制

仅仅封装 `HTTP Request` 是不够的，组件通过 `client-go` 和 `kube-apiserver` 交互，必然对实时性，可靠性等有很高要求。试想，如果 `ETCD` 中存储的数据和组件通过 `client-go` 从 `ETCD` 获取的数据不匹配的话，那将会是一个非常严重的问题。

如何实现 `client-go` 的实时性，可靠性？`client-go` 给出的答案是：`informer` 机制。

![client-go informer](img/client-go%20informer.png)  
&emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; *client-go informer 流程图*

`informer` 机制的核心组件包括：
- `Reflector`: 主要负责两类任务：
  1. 通过 `client-go` 客户端对象 list `kube-apiserver` 资源，并且 watch `kube-apiserver` 资源变更。
  2. 作为生产者，将获取的资源放入 `Delta FIFO` 队列。
- `Informer`: 主要负责三类任务：
  1. 作为消费者，将 `Reflector` 放入队列的资源拿出来。
  2. 将资源交给 `indexer` 组件。
  3. 交给 `indexer` 组件之后触发回调函数，处理回调事件。
- `Indexer`: `indexer` 组件负责将资源信息存入到本地内存数据库（实际是 `map` 对象），该数据库作为缓存存在，其资源信息和 `ETCD` 中的资源信息完全一致（得益于 `watch` 机制）。因此，`client-go` 可以从本地 `indexer` 中读取相应的资源，而不用每次都从 `kube-apiserver` 中获取资源信息。这也实现了 `client-go` 对于实时性的要求。

接下来从源码角度看各个组件的处理流程，力图做到知其然，知其所以然。

# 2 informer 源码分析

直接阅读 `informer` 源码是非常晦涩难懂的，这里通过 `informer` 的代码示例开始学习：
```
package main

import (
	"log"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
    // 解析 kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

    // 创建 ClientSet 客户端对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

    // 创建 sharedInformers
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)
    // 创建 informer
	informer := sharedInformers.Core().V1().Pods().Informer()

    // 创建 Event 回调 handler
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("New Pod Added to Store: %s", mObj.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oObj := oldObj.(v1.Object)
			nObj := newObj.(v1.Object)
			log.Printf("%s Pod Updated to %s", oObj.GetName(), nObj.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("Pod Deleted from Store: %s", mObj.GetName())
		},
	})

    // 运行 informer
	informer.Run(stopCh)
}
```

执行结果如下：
```
# go run informer.go 
2023/12/14 12:00:26 New Pod Added to Store: prometheus-alertmanager-0
2023/12/14 12:01:26 prometheus-alertmanager-0 Pod Updated to prometheus-alertmanager-0
```

上述代码示例分为三部分：创建 `informer`，创建 `informer` 的 `EventHandler`，运行 `informer`。下面，通过这三部分流程介绍 `client-go` 的核心组件。

## 2.1 创建 `informer`

创建 `informer` 分为两步。

1）创建工厂 `sharedInformerFactory`
```
// sharedInformers factory 
sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)

// client-go/informers/factory.go
func NewSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync)
}

func NewSharedInformerFactoryWithOptions(client kubernetes.Interface, defaultResync time.Duration, options ...SharedInformerOption) SharedInformerFactory {
	factory := &sharedInformerFactory{
		client:           client,
		namespace:        v1.NamespaceAll,
		defaultResync:    defaultResync,
		informers:        make(map[reflect.Type]cache.SharedIndexInformer),
		startedInformers: make(map[reflect.Type]bool),
		customResync:     make(map[reflect.Type]time.Duration),
	}

	// Apply all options
	for _, opt := range options {
		factory = opt(factory)
	}

	return factory
}
```

`sharedInformerFactory` 实现了 `SharedInformerFactory` 接口，该工厂负责创建 `informer`。

2）创建 `informer`

```
// 创建 informer
informer := sharedInformers.Core().V1().Pods().Informer()

// 调用 Core 方法
func (f *sharedInformerFactory) Core() core.Interface {
	return core.New(f, f.namespace, f.tweakListOptions)
}

func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &group{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// 调用 V1 方法
func (g *group) V1() v1.Interface {
	return v1.New(g.factory, g.namespace, g.tweakListOptions)
}

func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// 调用 Pods 方法
func (v *version) Pods() PodInformer {
	return &podInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
```

经过层层构建创建 `podInformer` 对象，该对象实现了 `PodInformer` 接口，调用接口的 `Informer` 方法创建 `informer` 对象：
```
func (f *podInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&corev1.Pod{}, f.defaultInformer)
}
```

`podInformer.Informer` 实际调用的是 `sharedInformerFactory.InformerFor`：
```
func (f *sharedInformerFactory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

    // 反射出资源对象 obj 的 type 
	informerType := reflect.TypeOf(obj)

    // 读取并判断资源对象的 informer
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}

	...

    // 调用 newFunc 创建 informer
	informer = newFunc(f.client, resyncPeriod)

    // 将 type:informer 加入到 factory 的 informers 中
	f.informers[informerType] = informer

	return informer
}
```

从 `InformerFor` 方法可以看出，`sharedInformerFactory` 的 share 体现在同一个资源类型共享 `informer`。

这么设计在于，每个 `informer` 包括一个 `Reflector`，`Reflector` 通过访问 `kube-apiserver` 实现 `ListAndWatch` 操作。共享 `informer` 实际是共享 `Reflector`，这种共享机制将减少 `Reflector` 对于 `kube-apiserver` 的访问，降低 `kube-apiserver` 的负载，节约资源。

继续看，创建 `informer` 的 `newFunc` 函数做了什么：
```
informer = newFunc(f.client, resyncPeriod)

// client-go/informers/core/v1/pod.go
func (f *podInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPodInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func NewFilteredPodInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().Pods(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().Pods(namespace).Watch(context.TODO(), options)
			},
		},
		&corev1.Pod{},
		resyncPeriod,
		indexers,
	)
}
```

`newFunc` 实际调用的是 `NewFilteredPodInformer` 函数，在函数内创建 `cache.ListAndWatch` 对象，对象中包括 `ListFunc` 和 `WatchFunc` 回调函数，回调函数内调用 `ClientSet` 实现 list 和 watch 资源对象。

继续看 `cache.NewSharedIndexInformer`：
```
// client-go/tools/cache/shared_informer.go
func NewSharedIndexInformer(lw ListerWatcher, exampleObject runtime.Object, defaultEventHandlerResyncPeriod time.Duration, indexers Indexers) SharedIndexInformer {
	return NewSharedIndexInformerWithOptions(
		lw,
		exampleObject,
		SharedIndexInformerOptions{
			ResyncPeriod: defaultEventHandlerResyncPeriod,
			Indexers:     indexers,
		},
	)
}

func NewSharedIndexInformerWithOptions(lw ListerWatcher, exampleObject runtime.Object, options SharedIndexInformerOptions) SharedIndexInformer {
	realClock := &clock.RealClock{}

	return &sharedIndexInformer{
		indexer:                         NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, options.Indexers),
		processor:                       &sharedProcessor{clock: realClock},
		listerWatcher:                   lw,
		objectType:                      exampleObject,
		objectDescription:               options.ObjectDescription,
		resyncCheckPeriod:               options.ResyncPeriod,
		defaultEventHandlerResyncPeriod: options.ResyncPeriod,
		clock:                           realClock,
		cacheMutationDetector:           NewCacheMutationDetector(fmt.Sprintf("%T", exampleObject)),
	}
}
```

在 `NewSharedIndexInformerWithOptions` 函数内创建 informer `sharedIndexInformer`。可以看到，`sharedIndexInformer` 内包括了 `indexer` 核心组件。

`informer` 创建完成。接下来为 `informer` 添加回调函数 `EventHandler`。

## 2.2 创建 `EventHandler`

代码实现如下：
```
informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: func(obj interface{}) {
        mObj := obj.(v1.Object)
        log.Printf("New Pod Added to Store: %s", mObj.GetName())
    },
    UpdateFunc: func(oldObj, newObj interface{}) {
        oObj := oldObj.(v1.Object)
        nObj := newObj.(v1.Object)
        log.Printf("%s Pod Updated to %s", oObj.GetName(), nObj.GetName())
    },
    DeleteFunc: func(obj interface{}) {
        mObj := obj.(v1.Object)
        log.Printf("Pod Deleted from Store: %s", mObj.GetName())
    },
})
```

创建 `EventHandler` 的 `handler` 中包括三种回调函数：`AddFunc`，`UpdateFunc` 和 `DeleteFunc`，三种回调函数分别在资源有增加，变更，删除时触发。  

在 `sharedIndexInformer.AddEventHandler` 内，将 `handler` 传递给 `sharedIndexInformer.AddEventHandlerWithResyncPeriod` 方法，该方法主要创建 `listener` 对象：
```
// client-go/tools/cache/shared_informer.go
func (s *sharedIndexInformer) AddEventHandler(handler ResourceEventHandler) (ResourceEventHandlerRegistration, error) {
	return s.AddEventHandlerWithResyncPeriod(handler, s.defaultEventHandlerResyncPeriod)
}

func (s *sharedIndexInformer) AddEventHandlerWithResyncPeriod(handler ResourceEventHandler, resyncPeriod time.Duration) (ResourceEventHandlerRegistration, error) {
    ...
	listener := newProcessListener(handler, resyncPeriod, determineResyncPeriod(resyncPeriod, s.resyncCheckPeriod), s.clock.Now(), initialBufferSize, s.HasSynced)

    if !s.started {
		return s.processor.addListener(listener), nil
	}
    ...
}

// client-go/tools/cache/shared_informer.go
func newProcessListener(handler ResourceEventHandler, requestedResyncPeriod, resyncPeriod time.Duration, now time.Time, bufferSize int, hasSynced func() bool) *processorListener {
	ret := &processorListener{
		nextCh:                make(chan interface{}),
		addCh:                 make(chan interface{}),
		handler:               handler,
		syncTracker:           &synctrack.SingleFileTracker{UpstreamHasSynced: hasSynced},
		pendingNotifications:  *buffer.NewRingGrowing(bufferSize),
		requestedResyncPeriod: requestedResyncPeriod,
		resyncPeriod:          resyncPeriod,
	}

	ret.determineNextResync(now)

	return ret
}

func (p *sharedProcessor) addListener(listener *processorListener) ResourceEventHandlerRegistration {
    ...

	p.listeners[listener] = true
    ...

	return listener
}
```

`listener` 对象包含通道 `addCh` 和 `nextCh`，以及 `handler` 等对象。最后将 `listener` 存入 `sharedIndexInformer.sharedProcessor` 中。

创建完 `informer` 的 `EventHandler`，接下来该运行 `informer` 了。

