# 0. 前言

`containerd` 是一个行业标准的容器运行时，其强调简单性、健壮性和可移植性。本文将从 `containerd` 的代码结构入手，查看 `containerd` 的启动注册流程。

# 1. 启动注册流程

## 1.1 containerd

首先以调试模式运行 `containerd`：
```
// containerd/cmd/containerd/main.go
package main

import (
	...
	_ "github.com/containerd/containerd/v2/cmd/containerd/builtins"
)

...
func main() {
	app := command.App()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "containerd: %s\n", err)
		os.Exit(1)
	}
}
```

在启动 `containerd` 时，导入匿名包 `github.com/containerd/containerd/v2/cmd/containerd/builtins` 注册插件。

接着，进入 `command.App()`:  
```
// containerd/cmd/containerd/server/server.go
func App() *cli.App {
    app := cli.NewApp()
	app.Name = "containerd"
    ...

    app.Action = func(context *cli.Context) error {
		...
        go func() {
			defer close(chsrv)

			server, err := server.New(ctx, config)
			if err != nil {
				select {
				case chsrv <- srvResp{err: err}:
				case <-ctx.Done():
				}
				return
			}
			...
		}()
        ...
    }
}
```

这里省略了一系列初始化过程，重点在 `server.New(ctx, config)`。
```
// containerd/cmd/containerd/server/server.go
func New(ctx context.Context, config *srvconfig.Config) (*Server, error) {
    ...
    // 将插件加载到 loaded 中
    loaded, err := LoadPlugins(ctx, config)
    if err != nil {
		return nil, err
	}
    ...
    serverOpts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainStreamInterceptor(
			streamNamespaceInterceptor,
			prometheusServerMetrics.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			unaryNamespaceInterceptor,
			prometheusServerMetrics.UnaryServerInterceptor(),
		),
	}
    ...
    var (
		grpcServer = grpc.NewServer(serverOpts...)
		tcpServer  = grpc.NewServer(tcpServerOpts...)

        grpcServices  []grpcService
		tcpServices   []tcpService
		ttrpcServices []ttrpcService

		s = &Server{
			prometheusServerMetrics: prometheusServerMetrics,
			grpcServer:              grpcServer,
			tcpServer:               tcpServer,
			ttrpcServer:             ttrpcServer,
			config:                  config,
		}
        ...
    )
    ...
    // 遍历插件
    for _, p := range loaded {
        ...
        result := p.Init(initContext)
        if err := initialized.Add(result); err != nil {
			return nil, fmt.Errorf("could not add plugin result to plugin set: %w", err)
		}

		instance, err := result.Instance()
        ...
        if src, ok := instance.(grpcService); ok {
			grpcServices = append(grpcServices, src)
		}
		if src, ok := instance.(ttrpcService); ok {
			ttrpcServices = append(ttrpcServices, src)
		}
		if service, ok := instance.(tcpService); ok {
			tcpServices = append(tcpServices, service)
		}
        ...
    }

    // 注册插件服务
	for _, service := range grpcServices {
		if err := service.Register(grpcServer); err != nil {
			return nil, err
		}
	}
	for _, service := range ttrpcServices {
		if err := service.RegisterTTRPC(ttrpcServer); err != nil {
			return nil, err
		}
	}
	for _, service := range tcpServices {
		if err := service.RegisterTCP(tcpServer); err != nil {
			return nil, err
		}
	}
    ...
}
```

`server.New` 是 `containerd` 运行的主逻辑。

首先，将注册的插件加载到 `loaded`，接着遍历 `loaded`。通过 `result := p.Init(initContext)` 获取插件的实例。  
以 `io.containerd.grpc.v1.containers` 插件为例，查看 `p.Init` 是如何获取插件对象的。  
```
// containerd/vendor/github.com/containerd/plugin/plugin.go
func (r Registration) Init(ic *InitContext) *Plugin {
    // 调用注册插件的 InitFn 函数
	p, err := r.InitFn(ic)
	return &Plugin{
		Registration: r,
		Config:       ic.Config,
		Meta:         *ic.Meta,
		instance:     p,
		err:          err,
	}
}

// containerd/plugins/services/containers/service.go
func init() {
	registry.Register(&plugin.Registration{
		Type: plugins.GRPCPlugin,
		ID:   "containers",
		Requires: []plugin.Type{
			plugins.ServicePlugin,
		},
        // 执行 InitFn 返回 service 对象
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			i, err := ic.GetByID(plugins.ServicePlugin, services.ContainersService)
			if err != nil {
				return nil, err
			}
			return &service{local: i.(api.ContainersClient)}, nil
		},
	})
}
```

获取到插件实例后，根据插件类型注册插件实例以提供对应的（grpc/ttrpc/tcp）服务。

## 1.2 注册插件

注册插件是通过 `init` 机制实现的。在 `main` 中导入 `github.com/containerd/containerd/v2/cmd/containerd/builtins` 包。

`builtins` 包导入包含 `init` 的插件包实现插件注册。以 `cri` 插件为例：  
```
// containerd/cmd/containerd/builtins/cri.go
package builtins

import (
	_ "github.com/containerd/containerd/v2/plugins/cri"
	...
)

// containerd/plugins/cri/cri.go
package cri

...
// Register CRI service plugin
func init() {
	defaultConfig := criconfig.DefaultServerConfig()
	registry.Register(&plugin.Registration{
		Type: plugins.GRPCPlugin,
		ID:   "cri",
		Requires: []plugin.Type{
			...
		},
		Config: &defaultConfig,
		ConfigMigration: func(ctx context.Context, configVersion int, pluginConfigs map[string]interface{}) error {
			...
		},
		InitFn: initCRIService,
	})
}
```

在 `init` 中通过 `registry.Register` 注册插件：  
```
package registry
...
var register = struct {
	sync.RWMutex
	r plugin.Registry
}{}

// Register allows plugins to register
func Register(r *plugin.Registration) {
	register.Lock()
	defer register.Unlock()
	register.r = register.r.Register(r)
}
```

可以看到插件注册的过程实际是将插件结构体 `plugin.Registration` 注册到 `register.plugin.Registry` 的过程。

`register.plugin.Registry` 实际是一个包含 `Registration` 的切片。
```
package plugin

type Registry []*Registration
```

## 1.3 查看插件

使用 `ctr` 查看 `containerd` 注册的插件，`ctr` 是 `containerd` 官方提供的命令行工具。如下：  
```
# ctr plugins ls
TYPE                                   ID                       PLATFORMS      STATUS
io.containerd.image-verifier.v1        bindir                   -              ok
io.containerd.internal.v1              opt                      -              ok
...
```

# 2. 小结

本文主要介绍了 `containerd` 的启动注册插件流程。当然，插件的类型众多，插件是如何工作的，插件之间如何交互，kubernetes 又是怎么和 `containerd` 交互的，这些会在下文中继续介绍。
