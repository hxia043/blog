接着 [Kubernetes:kube-apiserver 之启动流程(一)](https://www.cnblogs.com/xingzheanan/p/17787066.html) 加以介绍。

### 1.2.2 创建 APIExtensions Server

创建完通用 `APIServer` 后继续创建 `APIExtensions Server`。
```
func (c completedConfig) New(delegationTarget genericapiserver.DelegationTarget) (*CustomResourceDefinitions, error) {
	genericServer, err := c.GenericConfig.New("apiextensions-apiserver", delegationTarget)

	s := &CustomResourceDefinitions{
		GenericAPIServer: genericServer,
	}

    // 存储建立 REST API 到资源实体的信息
    apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apiextensions.GroupName, Scheme, metav1.ParameterCodec, Codecs)

    // 资源实体
	storage := map[string]rest.Storage{}

	// customresourcedefinitions
	if resource := "customresourcedefinitions"; apiResourceConfig.ResourceEnabled(v1.SchemeGroupVersion.WithResource(resource)) {
        // 创建资源实体
		customResourceDefinitionStorage, err := customresourcedefinition.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter)
		if err != nil {
			return nil, err
		}
		storage[resource] = customResourceDefinitionStorage
		storage[resource+"/status"] = customresourcedefinition.NewStatusREST(Scheme, customResourceDefinitionStorage)
	}
	if len(storage) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap[v1.SchemeGroupVersion.Version] = storage
	}

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}
```

`APIGroupInfo` 对象用于描述资源组信息，`storage` 存储资源到资源实体的对应关系。

资源实体，通过 `NewREST()` 函数创建。
```
# kubernetes/vendor/k8s.io/apiextensions-apiserver/pkg/registry/customresourcedefinition/etcd.go
package customresourcedefinition

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*REST, error) {
	strategy := NewStrategy(scheme)

	store := &genericregistry.Store{
		NewFunc:                   func() runtime.Object { return &apiextensions.CustomResourceDefinition{} },
		NewListFunc:               func() runtime.Object { return &apiextensions.CustomResourceDefinitionList{} },
		PredicateFunc:             MatchCustomResourceDefinition,
		DefaultQualifiedResource:  apiextensions.Resource("customresourcedefinitions"),
		SingularQualifiedResource: apiextensions.Resource("customresourcedefinition"),

		CreateStrategy:      strategy,
		UpdateStrategy:      strategy,
		DeleteStrategy:      strategy,
		ResetFieldsStrategy: strategy,

		// TODO: define table converter that exposes more than name/creation timestamp
		TableConvertor: rest.NewDefaultTableConvertor(apiextensions.Resource("customresourcedefinitions")),
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}
	return &REST{store}, nil
}
```

可以看到，资源实体是在资源包 `customresourcedefinition` 的 `etcd.go` 中创建的，创建的资源实体负责和 `etcd` 交互。  
(*关于 `etcd` 交互的部分先不讲，后续会专门介绍。*)

创建完资源实体后，通过 `apiGroupInfo.VersionedResourcesStorageMap[v1.SchemeGroupVersion.Version] = storage` 将资源实体存储到 `apiGroupInfo`。

继续调用 `InstallAPIGroup(apiGroupInfo *APIGroupInfo)` 安装 `REST API`。
```
# kubernetes/vendor/k8s.io/apiserver/pkg/server/genericapiserver.go
func (s *GenericAPIServer) InstallAPIGroup(apiGroupInfo *APIGroupInfo) error {
	return s.InstallAPIGroups(apiGroupInfo)
}

func (s *GenericAPIServer) InstallAPIGroups(apiGroupInfos ...*APIGroupInfo) error {
    for _, apiGroupInfo := range apiGroupInfos {
        // 调用 installAPIResources
		if err := s.installAPIResources(APIGroupPrefix, apiGroupInfo, openAPIModels); err != nil {
			return fmt.Errorf("unable to install api resources: %v", err)
		}
    }
    s.DiscoveryGroupManager.AddGroup(apiGroup)
	s.Handler.GoRestfulContainer.Add(discovery.NewAPIGroupHandler(s.Serializer, apiGroup).WebService())
}

// installAPIResources is a private method for installing the REST storage backing each api groupversionresource
func (s *GenericAPIServer) installAPIResources(apiPrefix string, apiGroupInfo *APIGroupInfo, typeConverter managedfields.TypeConverter) error {
	for _, groupVersion := range apiGroupInfo.PrioritizedVersions {
		apiGroupVersion, err := s.getAPIGroupVersion(apiGroupInfo, groupVersion, apiPrefix)
		if err != nil {
			return err
		}

        // 调用 InstallREST
		discoveryAPIResources, r, err := apiGroupVersion.InstallREST(s.Handler.GoRestfulContainer)
    }
}

# kubernetes/vendor/k8s.io/apiserver/pkg/endpoints/groupversion.go
func (g *APIGroupVersion) InstallREST(container *restful.Container) ([]apidiscoveryv2beta1.APIResourceDiscovery, []*storageversion.ResourceInfo, error) {
	prefix := path.Join(g.Root, g.GroupVersion.Group, g.GroupVersion.Version)
	installer := &APIInstaller{
		group:             g,
		prefix:            prefix,
		minRequestTimeout: g.MinRequestTimeout,
	}

    // 调用 Install
	apiResources, resourceInfos, ws, registrationErrors := installer.Install()
	container.Add(ws)

	return aggregatedDiscoveryResources, removeNonPersistedResources(resourceInfos), utilerrors.NewAggregate(registrationErrors)
}

# kubernetes/vendor/k8s.io/apiserver/pkg/endpoints/installer.go
// Install handlers for API resources.
func (a *APIInstaller) Install() ([]metav1.APIResource, []*storageversion.ResourceInfo, *restful.WebService, []error) {
	for _, path := range paths {
        // 注册资源 Handler
		apiResource, resourceInfo, err := a.registerResourceHandlers(path, a.group.Storage[path], ws)
		if err != nil {
			errors = append(errors, fmt.Errorf("error in registering resource: %s, %v", path, err))
		}
		if apiResource != nil {
			apiResources = append(apiResources, *apiResource)
		}
		if resourceInfo != nil {
			resourceInfos = append(resourceInfos, resourceInfo)
		}
	}
	return apiResources, resourceInfos, ws, errors
}
```

如上例所示，注册资源 `REST API` 的调用链很长，通过逐层调用，走到 `registerResourceHandlers` 注册资源 `handler`。

`registerResourceHandlers` 函数非常的长，主要抓一点：注册 `RESTful API` 的资源 `handler` 需要什么？回答好这个问题基本上就能抓住 `registerResourceHandlers` 的精髓了。

注册资源 `handler` 需要知道资源的 `API path` 和资源实体（*指和 etcd 交互的资源实例*）的对应关系，接着需要知道哪些 `action` 可以访问 `API path`。

围绕这两块看资源 `handler` 的注册过程。
```
func (a *APIInstaller) registerResourceHandlers(path string, storage rest.Storage, ws *restful.WebService) (*metav1.APIResource, *storageversion.ResourceInfo, error) {
	// what verbs are supported by the storage, used to know what verbs we support per path
	creater, isCreater := storage.(rest.Creater)
	namedCreater, isNamedCreater := storage.(rest.NamedCreater)
	lister, isLister := storage.(rest.Lister)
	getter, isGetter := storage.(rest.Getter)
	getterWithOptions, isGetterWithOptions := storage.(rest.GetterWithOptions)
	gracefulDeleter, isGracefulDeleter := storage.(rest.GracefulDeleter)
	collectionDeleter, isCollectionDeleter := storage.(rest.CollectionDeleter)
	updater, isUpdater := storage.(rest.Updater)
	patcher, isPatcher := storage.(rest.Patcher)
	watcher, isWatcher := storage.(rest.Watcher)
	connecter, isConnecter := storage.(rest.Connecter)
	storageMeta, isMetadata := storage.(rest.StorageMetadata)
	storageVersionProvider, isStorageVersionProvider := storage.(rest.StorageVersionProvider)
	gvAcceptor, _ := storage.(rest.GroupVersionAcceptor)


	// Get the list of actions for the given scope.
	switch {
	case !namespaceScoped:
		...
	default:
		// Handler for standard REST verbs (GET, PUT, POST and DELETE).
		actions := []action{}

		actions = appendIf(actions, action{"LIST", resourcePath, resourceParams, namer, false}, isLister)
		actions = appendIf(actions, action{"POST", resourcePath, resourceParams, namer, false}, isCreater)
		actions = appendIf(actions, action{"DELETECOLLECTION", resourcePath, resourceParams, namer, false}, isCollectionDeleter)
		// DEPRECATED in 1.11
		actions = appendIf(actions, action{"WATCHLIST", "watch/" + resourcePath, resourceParams, namer, false}, allowWatchList)

		actions = appendIf(actions, action{"GET", itemPath, nameParams, namer, false}, isGetter)
		if getSubpath {
			actions = appendIf(actions, action{"GET", itemPath + "/{path:*}", proxyParams, namer, false}, isGetter)
		}
		actions = appendIf(actions, action{"PUT", itemPath, nameParams, namer, false}, isUpdater)
		actions = appendIf(actions, action{"PATCH", itemPath, nameParams, namer, false}, isPatcher)
		actions = appendIf(actions, action{"DELETE", itemPath, nameParams, namer, false}, isGracefulDeleter)
		// DEPRECATED in 1.11
		actions = appendIf(actions, action{"WATCH", "watch/" + itemPath, nameParams, namer, false}, isWatcher)
		actions = appendIf(actions, action{"CONNECT", itemPath, nameParams, namer, false}, isConnecter)
		actions = appendIf(actions, action{"CONNECT", itemPath + "/{path:*}", proxyParams, namer, false}, isConnecter && connectSubpath)
	}

	for _, action := range actions {
		switch action.Verb {
		case "GET": // Get a resource.
			var handler restful.RouteFunction
			if isGetterWithOptions {
				handler = restfulGetResourceWithOptions(getterWithOptions, reqScope, isSubresource)
			} else {
				handler = restfulGetResource(getter, reqScope)
			}

			route := ws.GET(action.Path).To(handler).
				Doc(doc).
				Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
				Operation("read"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(append(storageMeta.ProducesMIMETypes(action.Verb), mediaTypes...)...).
				Returns(http.StatusOK, "OK", producedObject).
				Writes(producedObject)

			routes = append(routes, route)
		}

		for _, route := range routes {
			route.Metadata(ROUTE_META_GVK, metav1.GroupVersionKind{
				Group:   reqScope.Kind.Group,
				Version: reqScope.Kind.Version,
				Kind:    reqScope.Kind.Kind,
			})
			route.Metadata(ROUTE_META_ACTION, strings.ToLower(action.Verb))
			ws.Route(route)
		}
	}
}
```

可以看到，通过资源实体 `storage` 支持的接口类型可以反射出资源支持的方法。接着将支持的方法加入 `actions`。`actions` 存储的是 `action.Verb` 和支持的资源 `API path` 信息。

拿到 `actions` 后，通过匹配 `actions.Verb` 建立 `action.Verb -> action.Path -> handler` 的路由。其中，创建 `handler` 需要用到 `storage`，因为 `handler` 是通过 `storage` 和 `etcd` 交互的。

通过 `go-restful` 库建立上述路由，接着 `ws.Route(route)` 将 `route` 加入到 `restful.WebService` 中。

回头看 `InstallREST`。
```
func (g *APIGroupVersion) InstallREST(container *restful.Container) ([]apidiscoveryv2beta1.APIResourceDiscovery, []*storageversion.ResourceInfo, error) {

	apiResources, resourceInfos, ws, registrationErrors := installer.Install()
	container.Add(ws)

	return aggregatedDiscoveryResources, removeNonPersistedResources(resourceInfos), utilerrors.NewAggregate(registrationErrors)
}

func (s *GenericAPIServer) installAPIResources(apiPrefix string, apiGroupInfo *APIGroupInfo, typeConverter managedfields.TypeConverter) error {
	discoveryAPIResources, r, err := apiGroupVersion.InstallREST(s.Handler.GoRestfulContainer)
}
```

通过 `container.Add(ws)` 将 `ws` 添加到 `go-restful` 的 `container` 中，该 `container` 即为 `APIExtensions Server` 的 `Handler` 所提供的 `GoRestfulContainer`。

这里介绍了 `APIExtensions Server` 的 `REST API` 创建过程。对于 `KubeAPIServer` 和 `AggregatorServer` 的创建过程与之类似，不过多介绍。

`REST API` 创建好以后，下一步就到如何运行 `kube-apiserver`了。

## 1.3 运行 kube-apiserver

`kube-apiserver` 作为提供 `RESTful API` 的组件，其运行主要是监听端口和启动服务。理清了这点，就能在复杂的运行代码中找出头绪。

调用 `APIAggregator.PrepareRun` 和 `preparedAPIAggregator.Run` 运行 `kube-apiserver`。
```
func Run(opts options.CompletedOptions, stopCh <-chan struct{}) error {
	prepared, err := server.PrepareRun()
	if err != nil {
		return err
	}

	return prepared.Run(stopCh)
}
```

启动过程在 `PrepareRun` 中。
```
func (s *APIAggregator) PrepareRun() (preparedAPIAggregator, error) {
	prepared := s.GenericAPIServer.PrepareRun()
	return preparedAPIAggregator{APIAggregator: s, runnable: prepared}, nil
}
```

运行 `prepared.Run(stopCh)` 实际调用的是 `preparedGenericAPIServer.Run` 方法。
```
func (s preparedGenericAPIServer) Run(stopCh <-chan struct{}) error {
	// 调用 preparedGenericAPIServer.NonBlockingRun
	stoppedCh, listenerStoppedCh, err := s.NonBlockingRun(stopHttpServerCh, shutdownTimeout)
	if err != nil {
		return err
	}
}

func (s preparedGenericAPIServer) NonBlockingRun(stopCh <-chan struct{}, shutdownTimeout time.Duration) (<-chan struct{}, <-chan struct{}, error) {
	if s.SecureServingInfo != nil && s.Handler != nil {
		var err error
		// 调用 SecureServingInfo.Serve
		stoppedCh, listenerStoppedCh, err = s.SecureServingInfo.Serve(s.Handler, shutdownTimeout, internalStopCh)
		if err != nil {
			close(internalStopCh)
			return nil, nil, err
		}
	}
}

func (s *SecureServingInfo) Serve(handler http.Handler, shutdownTimeout time.Duration, stopCh <-chan struct{}) (<-chan struct{}, <-chan struct{}, error) {
	// 组 http.Server
	secureServer := &http.Server{
		Addr:           s.Listener.Addr().String(),
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig,

		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second, // just shy of requestTimeoutUpperBound
	}

	return RunServer(secureServer, s.Listener, shutdownTimeout, stopCh)
}

func RunServer(
	server *http.Server,
	ln net.Listener,
	shutDownTimeout time.Duration,
	stopCh <-chan struct{},
) (<-chan struct{}, <-chan struct{}, error) {
	go func() {
		// 调用 http 的 Server.Serve 提供 `RESTful API` 服务
		err := server.Serve(listener)

		msg := fmt.Sprintf("Stopped listening on %s", ln.Addr().String())
	}
}
```

可以看到，最终调用 `http` 包的 `Server.Serve` 提供 `RESTful API` 服务。

至此，已介绍完 `kube-apiserver` 从启动到运行的核心逻辑。下一篇，将重点介绍 `kube-apiserver` 是怎么和 `etcd` 进行交互的。
