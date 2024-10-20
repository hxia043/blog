# 0. 前言

[Kubernetes:kubelet 源码分析之创建 pod 流程](https://www.cnblogs.com/xingzheanan/p/18202067) 介绍了 `kubelet` 创建 pod 的流程，其中介绍了 `kubelet` 调用 runtime cri 接口创建 pod。[containerd 源码分析：启动注册流程](https://www.cnblogs.com/xingzheanan/p/18204637) 介绍了 `containerd` 作为一种行业标准的高级运行时的启动注册流程。那么，`kubelet` 是怎么和 `containerd` 交互的呢? 本文会带着这个问题分析 `kubelet` 和 `containerd` 的交互。

# 1. kubelet 和 containerd 交互

## 1.1 kubelet

如 [Kubernetes:kubelet 源码分析之创建 pod 流程](https://www.cnblogs.com/xingzheanan/p/18202067) 分析，`kubelet` 调用 runtime cri 接口 `/runtime.v1.RuntimeService/RunPodSandbox` 创建 pod：
```
// kubernetes/vendor/k8s.io/cri-api/pkg/apis/runtime/v1/api.pb.go
func (c *runtimeServiceClient) RunPodSandbox(ctx context.Context, in *RunPodSandboxRequest, opts ...grpc.CallOption) (*RunPodSandboxResponse, error) {
	out := new(RunPodSandboxResponse)
	err := c.cc.Invoke(ctx, "/runtime.v1.RuntimeService/RunPodSandbox", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
```

`kubelet` 是 runtime cri 接口调用的客户端，那么容器运行时作为服务端是怎么提供服务的呢？

## 1.2 kubelet 和 containerd 交互流程

在介绍容器运行时提供的服务之前先看下 [cri 架构图](https://github.com/containerd/containerd/blob/v1.7.0/docs/cri/architecture.md)。

![kubelet 和 containerd 架构图](./img/containerd%20和%20pod%20交互流程图.png)

从图中可以看出，`containerd` 的 CRI 插件提供 `image service` 和 `runtime service`，负责对接 `kubelet` runtime cri 的接口调用，并将调用转发给 `containerd`。

继续，查看 `containerd` 的处理流程。

## 1.3 containerd

### 1.3.1 CRI Plugin

根据 `cri 架构图`, 从 `CRI` 插件入手查看 id 为 `io.containerd.grpc.v1.cri` 的 `CRI` 插件。
```
// containerd/plugins/cri/cri.go
func initCRIService(ic *plugin.InitContext) (interface{}, error) {
	...
    // Get runtime service.
	criRuntimePlugin, err := ic.GetByID(plugins.CRIServicePlugin, "runtime")
	if err != nil {
		return nil, fmt.Errorf("unable to load CRI runtime service plugin dependency: %w", err)
	}

    // Get image service.
	criImagePlugin, err := ic.GetByID(plugins.CRIServicePlugin, "images")
	if err != nil {
		return nil, fmt.Errorf("unable to load CRI image service plugin dependency: %w", err)
	}
    ...
    service := &criGRPCServer{
		RuntimeServiceServer: rs,
		ImageServiceServer:   is,
		Closer:               s, // TODO: Where is close run?
		initializer:          s,
	}

	if config.DisableTCPService {
		return service, nil
	}

	return criGRPCServerWithTCP{service}, nil
}
```

插件返回的是 `criGRPCServerWithTCP` 对象。其中，包括 `criGRPCServer` 对象。`criGRPCServer` 对象实现了 `grpcService` 接口，将调用接口的 `Register` 注册对象到 grpc server。
```
// containerd/plugins/cri/cri.go
// Register registers all required services onto a specific grpc server.
// This is used by containerd cri plugin.
func (c *criGRPCServer) Register(s *grpc.Server) error {
	return c.register(s)
}

func (c *criGRPCServer) register(s *grpc.Server) error {
	instrumented := instrument.NewService(c)
	runtime.RegisterRuntimeServiceServer(s, instrumented)
	runtime.RegisterImageServiceServer(s, instrumented)
	return nil
}
```

在 `criGRPCServer.register` 中创建 `instrumentedService` 对象。
```
type instrumentedService struct {
	c criService
}

func NewService(c criService) GRPCServices {
	return &instrumentedService{c: c}
}
```

`instrumentedService` 包括 `criService` 对象。实际提供 `runtime service` 和 `image service` 的就是 `criService` 对象。

以注册 `runtime service` 为例，查看 `runtime.RegisterRuntimeServiceServer(s, instrumented)` 做了什么。
```
// containerd/vendor/k8s.io/cri-api/pkg/apis/runtime/v1/api.pb.go
func RegisterRuntimeServiceServer(s *grpc.Server, srv RuntimeServiceServer) {
	s.RegisterService(&_RuntimeService_serviceDesc, srv)
}

var _RuntimeService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "runtime.v1.RuntimeService",
	HandlerType: (*RuntimeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _RuntimeService_Version_Handler,
		},
		{
			MethodName: "RunPodSandbox",
			Handler:    _RuntimeService_RunPodSandbox_Handler,
		},
        ...
    },
    ...
}
```

可以看到，注册 `instrumentedService` 到 grpc 中，`instrumentedService` 提供 `runtime.v1.RuntimeService` 服务，包括 `kubelet` 调用的 `RunPodSandbox` 方法。

继续看，`instrumentedService` 的 `RunPodSandbox` 做了什么。
```
// containerd/internal/cri/instrumented_service.go
func (in *instrumentedService) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (res *runtime.RunPodSandboxResponse, err error) {
	...
	res, err = in.c.RunPodSandbox(ctrdutil.WithNamespace(ctx), r)
	return res, errdefs.ToGRPC(err)
}
```

`instrumentedService` 调用 `criGRPCServer` 的 `RunPodSandbox` 方法，实际执行的是 `criGRPCServer` 中的 `criServer` 对象：
```
// containerd/internal/cri/server/sandbox_run.go
func (c *criService) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (_ *runtime.RunPodSandboxResponse, retErr error) {
	...
    if err := c.sandboxService.CreateSandbox(ctx, sandboxInfo, sb.WithOptions(config), sb.WithNetNSPath(sandbox.NetNSPath)); err != nil {
		return nil, fmt.Errorf("failed to create sandbox %q: %w", id, err)
	}

	ctrl, err := c.sandboxService.StartSandbox(ctx, sandbox.Sandboxer, id)
    if err != nil {
        ...
    }
    ...
}
```

`criService.RunPodSandbox` 调用的是 `sandboxService` 的 `CreateSandbox` 和 `StartSandbox` 方法。

### 1.3.2 sanbox Plugin

`sandboxService` 在 `cri.initCRIService` 中实例化：
```
// containerd/plugins/cri/cri.go
func initCRIService(ic *plugin.InitContext) (interface{}, error) {
    ...
    sbControllers, err := getSandboxControllers(ic)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox controllers from plugins %v", err)
	}
    ...
    options := &server.CRIServiceOptions{
		RuntimeService:     criRuntimePlugin.(server.RuntimeService),
		ImageService:       criImagePlugin.(server.ImageService),
		StreamingConfig:    streamingConfig,
		NRI:                getNRIAPI(ic),
		Client:             client,
		SandboxControllers: sbControllers,
	}
    ...
    s, rs, err := server.NewCRIService(options)
    ...
    service := &criGRPCServer{
		RuntimeServiceServer: rs,
		ImageServiceServer:   is,
		Closer:               s, // TODO: Where is close run?
		initializer:          s,
	}
}
```

首先，`getSandboxControllers` 获得 `sandbox controllers`:
```
// containerd/plugins/cri/cri.go
func getSandboxControllers(ic *plugin.InitContext) (map[string]sandbox.Controller, error) {
    // plugins.SandboxControllerPlugin: "io.containerd.sandbox.controller.v1"
	sandboxers, err := ic.GetByType(plugins.SandboxControllerPlugin)
	if err != nil {
		return nil, err
	}
	...
	return sc, nil
}
```

`sandbox.Controller` 是类型为 `io.containerd.sandbox.controller.v1` 的插件对象。将该对象作为 `options` 赋给 `criServer`：
```
// containerd/internal/cri/server/service.go
func NewCRIService(options *CRIServiceOptions) (CRIService, runtime.RuntimeServiceServer, error) {
	...
	c := &criService{
		...
		sandboxService:     newCriSandboxService(&config, options.SandboxControllers),
	}
    ...
}

func newCriSandboxService(config *criconfig.Config, sandboxers map[string]sandbox.Controller) *criSandboxService {
	return &criSandboxService{
		sandboxControllers: sandboxers,
		config:             config,
	}
}
```

`criService.sandboxService.CreateSandbox` 调用的是插件对象 `sanbox controllers` 的 `CreateSandbox` 方法，该方法最终调用的是 `sandboxClient` 的 `CreateSandbox`：
```
func (c *sandboxClient) CreateSandbox(ctx context.Context, in *CreateSandboxRequest, opts ...grpc.CallOption) (*CreateSandboxResponse, error) {
	out := new(CreateSandboxResponse)
	err := c.cc.Invoke(ctx, "/containerd.runtime.sandbox.v1.Sandbox/CreateSandbox", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
```

可以看到，在 `sandboxClient.CreateSandbox` 中调用 `containerd` 提供的 containerd cri 接口 `/containerd.runtime.sandbox.v1.Sandbox/CreateSandbox`，该接口用来创建 sandbox，即 pod。

## 1.4 创建 pod 流程

根据上述分析，这里画出 `kubelet` 到 `containerd` 的交互流程图如下：  
![kubelet 和 containerd 交互流程](./img/kubelet%20和%20contaienrd%20交互.png)

# 2. 小结

本文在前文 [Kubernetes:kubelet 源码分析之创建 pod 流程](https://www.cnblogs.com/xingzheanan/p/18202067) 和 [containerd 源码分析：启动注册流程](https://www.cnblogs.com/xingzheanan/p/18204637) 的基础上，进一步分析从 `kubelet` 到 `containerd` 的交互流程，打通了 `kubelet` 到 `containerd` 这一步。
