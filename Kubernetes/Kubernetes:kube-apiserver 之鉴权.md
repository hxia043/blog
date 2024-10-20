`kubernetes:kube-apiserver` 系列文章：
- [Kubernetes:kube-apiserver 之 scheme(一)](https://www.cnblogs.com/xingzheanan/p/17771090.html)
- [Kubernetes:kube-apiserver 之 scheme(二)](https://www.cnblogs.com/xingzheanan/p/17774196.html)
- [Kubernetes:kube-apiserver 之启动流程(一)](https://www.cnblogs.com/xingzheanan/p/17787066.html)
- [Kubernetes:kube-apiserver 之启动流程(二)](https://www.cnblogs.com/xingzheanan/p/17810006.html)
- [Kubernetes:kube-apiserver 和 etcd 的交互 ](https://www.cnblogs.com/xingzheanan/p/17810847.html)
- [Kubernetes:kube-apiserver 之认证](https://www.cnblogs.com/xingzheanan/p/17818588.html)

# 0. 前言

上一篇文章介绍了 `kube-apiserver` 的认证机制。这里继续往下走，介绍 `kube-apiserver` 的鉴权。`kube-apiserver` 处理认证和鉴权非常类似，建议阅读鉴权机制前先看看 `kube-apiserver` 的 [认证](https://www.cnblogs.com/xingzheanan/p/17818588.html)。

# 1. 鉴权 Authorization

## 1.1 Authorization handler

进入 `DefaultBuildHandlerChain` 看 `Authorization handler` 创建过程。
```
# kubernetes/vendor/k8s.io/apiserver/pkg/server/config.go
func DefaultBuildHandlerChain(apiHandler http.Handler, c *Config) http.Handler {
	handler := apiHandler

	handler = genericapifilters.WithAuthorization(handler, c.Authorization.Authorizer, c.Serializer)
	handler = filterlatency.TrackStarted(handler, c.TracerProvider, "authorization")

    handler = genericapifilters.WithAuthentication(handler, c.Authentication.Authenticator, failedHandler, c.Authentication.APIAudiences, c.Authentication.RequestHeaderConfig)
	handler = filterlatency.TrackStarted(handler, c.TracerProvider, "authentication")
}
```

这里 `handler chain` 的 `handler` 处理顺序是由下往上的，即处理完 `authentication handler` 在处理 `authorization handler`。

进入 `genericapifilters.WithAuthorization` 查看鉴权 `handler` 的创建过程。
```
# kubernetes/vendor/k8s.io/apiserver/pkg/endpoints/filters/authorization.go
func WithAuthorization(hhandler http.Handler, auth authorizer.Authorizer, s runtime.NegotiatedSerializer) http.Handler {
	return withAuthorization(hhandler, auth, s, recordAuthorizationMetrics)
}

func withAuthorization(handler http.Handler, a authorizer.Authorizer, s runtime.NegotiatedSerializer, metrics recordAuthorizationMetricsFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		authorizationStart := time.Now()

        // 获取 request 携带的用户信息
		attributes, err := GetAuthorizerAttributes(ctx)
		if err != nil {
			responsewriters.InternalError(w, req, err)
			return
		}

        // 对用户信息进行鉴权
		authorized, reason, err := a.Authorize(ctx, attributes)

		...
	})
}
```

鉴权过程包括两部分。  

一部分通过 `GetAuthorizerAttributes` 获取 `RESTful API` 请求中携带的用户信息。
```
# kubernetes/vendor/k8s.io/apiserver/pkg/endpoints/filters/authorization.go
func GetAuthorizerAttributes(ctx context.Context) (authorizer.Attributes, error) {
	attribs := authorizer.AttributesRecord{}

	user, ok := request.UserFrom(ctx)
	if ok {
		attribs.User = user
	}

	requestInfo, found := request.RequestInfoFrom(ctx)
	if !found {
		return nil, errors.New("no RequestInfo found in the context")
	}

	// Start with common attributes that apply to resource and non-resource requests
	attribs.ResourceRequest = requestInfo.IsResourceRequest
	attribs.Path = requestInfo.Path
	attribs.Verb = requestInfo.Verb

	attribs.APIGroup = requestInfo.APIGroup
	attribs.APIVersion = requestInfo.APIVersion
	attribs.Resource = requestInfo.Resource
	attribs.Subresource = requestInfo.Subresource
	attribs.Namespace = requestInfo.Namespace
	attribs.Name = requestInfo.Name

	return &attribs, nil
}
```

获取到用户信息后通过 `a.Authorize(ctx, attributes)` 对用户及其请求进行鉴权。其中 `a` 是实现了 `Authorizer` 鉴权器接口的实例。
```
type Authorizer interface {
	Authorize(ctx context.Context, a Attributes) (authorized Decision, reason string, err error)
}
```

查看 `a.Authorize(ctx, attributes)` 实际是看 `Config.Authorization.Authorizer` 中的实例。
```
func DefaultBuildHandlerChain(apiHandler http.Handler, c *Config) http.Handler {
	handler = genericapifilters.WithAuthorization(handler, c.Authorization.Authorizer, c.Serializer)
}
```

## 1.2 Authorization authorizer

`Config.Authorization.Authorizer` 在 `BuildGenericConfig` 的 `BuildAuthorizer` 函数内创建。
```
# kubernetes/pkg/controlplane/apiserver/config.go
func BuildGenericConfig(
	s controlplaneapiserver.CompletedOptions,
	schemes []*runtime.Scheme,
	getOpenAPIDefinitions func(ref openapicommon.ReferenceCallback) map[string]openapicommon.OpenAPIDefinition,
) (
	genericConfig *genericapiserver.Config,
	versionedInformers clientgoinformers.SharedInformerFactory,
	storageFactory *serverstorage.DefaultStorageFactory,

	lastErr error,
) {
    genericConfig.Authorization.Authorizer, genericConfig.RuleResolver, err = BuildAuthorizer(s, genericConfig.EgressSelector, versionedInformers)
	if err != nil {
		lastErr = fmt.Errorf("invalid authorization config: %v", err)
		return
	}
}
```

进入 `BuildAuthorizer` 查看 `Authorizer` 是怎么创建的。
```
# kubernetes/pkg/controlplane/apiserver/config.go
func BuildAuthorizer(s controlplaneapiserver.CompletedOptions, EgressSelector *egressselector.EgressSelector, versionedInformers clientgoinformers.SharedInformerFactory) (authorizer.Authorizer, authorizer.RuleResolver, error) {
	authorizationConfig := s.Authorization.ToAuthorizationConfig(versionedInformers)

	if EgressSelector != nil {
		egressDialer, err := EgressSelector.Lookup(egressselector.ControlPlane.AsNetworkContext())
		if err != nil {
			return nil, nil, err
		}
		authorizationConfig.CustomDial = egressDialer
	}

	return authorizationConfig.New()
}
```

创建 `Authorizer` 分为两块。首先创建 `authorizationConfig`，接着通过 `authorizationConfig.New()` 实例化 `authorizer`。
```
# kubernetes/pkg/kubeapiserver/authorizer/config.go
func (config Config) New() (authorizer.Authorizer, authorizer.RuleResolver, error) {
	var (
		authorizers   []authorizer.Authorizer
		ruleResolvers []authorizer.RuleResolver
	)

	for _, authorizationMode := range config.AuthorizationModes {
		switch authorizationMode {
		case modes.ModeNode:
			...
		case modes.ModeAlwaysAllow:
			...
		case modes.ModeAlwaysDeny:
            ...
		case modes.ModeABAC:
			...
		case modes.ModeWebhook:
			...
		case modes.ModeRBAC:
			...
		default:
			return nil, nil, fmt.Errorf("unknown authorization mode %s specified", authorizationMode)
		}
	}

	return union.New(authorizers...), union.NewRuleResolvers(ruleResolvers...), nil
}
```

可以看到，`authorizationConfig.New()` 内根据 `config.AuthorizationModes` 确定需要创建的鉴权器类型。这里有 `modes.ModeNode`，`modes.ModeAlwaysAllow`，`modes.ModeAlwaysDeny`，`modes.ModeABAC`，`modes.ModeWebhook` 和 `modes.ModeRBAC` 六种鉴权器。

`config.AuthorizationModes` 是在创建选项 `NewOptions` 中定义的，实例化过程如下：
```
func (o *BuiltInAuthorizationOptions) ToAuthorizationConfig(versionedInformerFactory versionedinformers.SharedInformerFactory) authorizer.Config {
	return authorizer.Config{
		AuthorizationModes:          o.Modes,
	}
}

# kubernetes/pkg/controlplane/apiserver/options/options.go
func NewOptions() *Options {
	s := Options{
		Authorization:           kubeoptions.NewBuiltInAuthorizationOptions()
	}

	return &s
}

# kubernetes/pkg/kubeapiserver/options/authorization.go
func NewBuiltInAuthorizationOptions() *BuiltInAuthorizationOptions {
	return &BuiltInAuthorizationOptions{
		Modes:                       []string{authzmodes.ModeAlwaysAllow},
		WebhookVersion:              "v1beta1",
		WebhookCacheAuthorizedTTL:   5 * time.Minute,
		WebhookCacheUnauthorizedTTL: 30 * time.Second,
		WebhookRetryBackoff:         genericoptions.DefaultAuthWebhookRetryBackoff(),
	}
}
```

这里的 `config.AuthorizationModes` 为 `authzmodes.ModeAlwaysAllow`。那么，将创建 `alwaysAllowAuthorizer` 鉴权器。
```
# kubernetes/pkg/kubeapiserver/authorizer/config.go
func (config Config) New() (authorizer.Authorizer, authorizer.RuleResolver, error) {
	var (
		authorizers   []authorizer.Authorizer
		ruleResolvers []authorizer.RuleResolver
	)

	for _, authorizationMode := range config.AuthorizationModes {
		switch authorizationMode {
		case modes.ModeAlwaysAllow:
			alwaysAllowAuthorizer := authorizerfactory.NewAlwaysAllowAuthorizer()
			authorizers = append(authorizers, alwaysAllowAuthorizer)
			ruleResolvers = append(ruleResolvers, alwaysAllowAuthorizer)
        }
    }

    return union.New(authorizers...), union.NewRuleResolvers(ruleResolvers...), nil
}

# kubernetes/vendor/k8s.io/apiserver/pkg/authorization/authorizerfactory/builtin.go
func NewAlwaysAllowAuthorizer() *alwaysAllowAuthorizer {
	return new(alwaysAllowAuthorizer)
}

type alwaysAllowAuthorizer struct{}

func (alwaysAllowAuthorizer) Authorize(ctx context.Context, a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	return authorizer.DecisionAllow, "", nil
}

func (alwaysAllowAuthorizer) RulesFor(user user.Info, namespace string) ([]authorizer.ResourceRuleInfo, []authorizer.NonResourceRuleInfo, bool, error) {
	return []authorizer.ResourceRuleInfo{
			&authorizer.DefaultResourceRuleInfo{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
			},
		}, []authorizer.NonResourceRuleInfo{
			&authorizer.DefaultNonResourceRuleInfo{
				Verbs:           []string{"*"},
				NonResourceURLs: []string{"*"},
			},
		}, false, nil
}
```

`alwaysAllowAuthorizer` 鉴权器实现了 `Authorizer` 接口，其总是返回 `authorizer.DecisionAllow`。

每种鉴权器通过 `union.New` 加到鉴权器组中。
```
func New(authorizationHandlers ...authorizer.Authorizer) authorizer.Authorizer {
	return unionAuthzHandler(authorizationHandlers)
}

// Authorizes against a chain of authorizer.Authorizer objects and returns nil if successful and returns error if unsuccessful
func (authzHandler unionAuthzHandler) Authorize(ctx context.Context, a authorizer.Attributes) (authorizer.Decision, string, error) {
	var (
		errlist    []error
		reasonlist []string
	)

	for _, currAuthzHandler := range authzHandler {
		decision, reason, err := currAuthzHandler.Authorize(ctx, a)

		if err != nil {
			errlist = append(errlist, err)
		}
		if len(reason) != 0 {
			reasonlist = append(reasonlist, reason)
		}
		switch decision {
		case authorizer.DecisionAllow, authorizer.DecisionDeny:
			return decision, reason, err
		case authorizer.DecisionNoOpinion:
			// continue to the next authorizer
		}
	}

	return authorizer.DecisionNoOpinion, strings.Join(reasonlist, "\n"), utilerrors.NewAggregate(errlist)
}
```

前面 `handler` 的 `a.Authorize(ctx, attributes)` 实际执行的是鉴权器组 `unionAuthzHandler` 的 `Authorize` 方法。在 `unionAuthzHandler.Authorize` 内遍历执行每种鉴权器的 `Authorize` 方法，如果有一种鉴权器鉴权通过，则返回鉴权成功。如果鉴权器返回 `authorizer.DecisionNoOpinion` 则执行下一个鉴权器。如果鉴权器鉴权失败则返回鉴权失败。

## 1.3 authorization rules

前面介绍 `alwaysAllowAuthorizer` 鉴权器的时候我们看到 `alwaysAllowAuthorizer.RulesFor` 方法，该方法内定义了用户可以访问的 `RESTful API` 资源和非 `RESTful API` 资源。如 `RESTful API` 资源定义了访问资源的动作 `Verbs`，资源组`APIGroups` 和资源 `Resources`。

上例的 `alwaysAllowAuthorizer` 并没有看出 `RulesFor` 的运用是因为 `alwaysAllowAuthorizer` 总是允许请求访问 `RESTful API` 资源和非 `RESTful API` 资源。

我们以 `rbacAuthorizer` 鉴权器为例，看 `RulesFor` 是怎么被用上的。
```
func (config Config) New() (authorizer.Authorizer, authorizer.RuleResolver, error) {
	for _, authorizationMode := range config.AuthorizationModes {
		// Keep cases in sync with constant list in k8s.io/kubernetes/pkg/kubeapiserver/authorizer/modes/modes.go.
		switch authorizationMode {
            case modes.ModeRBAC:
			rbacAuthorizer := rbac.New(
				&rbac.RoleGetter{Lister: config.VersionedInformerFactory.Rbac().V1().Roles().Lister()},
				&rbac.RoleBindingLister{Lister: config.VersionedInformerFactory.Rbac().V1().RoleBindings().Lister()},
				&rbac.ClusterRoleGetter{Lister: config.VersionedInformerFactory.Rbac().V1().ClusterRoles().Lister()},
				&rbac.ClusterRoleBindingLister{Lister: config.VersionedInformerFactory.Rbac().V1().ClusterRoleBindings().Lister()},
			)
			authorizers = append(authorizers, rbacAuthorizer)
			ruleResolvers = append(ruleResolvers, rbacAuthorizer)
        }
    }
}

func New(roles rbacregistryvalidation.RoleGetter, roleBindings rbacregistryvalidation.RoleBindingLister, clusterRoles rbacregistryvalidation.ClusterRoleGetter, clusterRoleBindings rbacregistryvalidation.ClusterRoleBindingLister) *RBACAuthorizer {
	authorizer := &RBACAuthorizer{
		authorizationRuleResolver: rbacregistryvalidation.NewDefaultRuleResolver(
			roles, roleBindings, clusterRoles, clusterRoleBindings,
		),
	}
	return authorizer
}

func (r *RBACAuthorizer) Authorize(ctx context.Context, requestAttributes authorizer.Attributes) (authorizer.Decision, string, error) {
	ruleCheckingVisitor := &authorizingVisitor{requestAttributes: requestAttributes}

	r.authorizationRuleResolver.VisitRulesFor(requestAttributes.GetUser(), requestAttributes.GetNamespace(), ruleCheckingVisitor.visit)
	if ruleCheckingVisitor.allowed {
		return authorizer.DecisionAllow, ruleCheckingVisitor.reason, nil
	}
}
```

可以看到，`rbacAuthorizer` 鉴权器的 `Authorize` 方法内的 `r.authorizationRuleResolver.VisitRulesFor` 结合用户信息和鉴权器的 `rules` 判断鉴权是否通过。

# 2. 总结

通过本篇文章介绍了 `kube-apiserver` 中的 `Authorization` 鉴权流程，下一篇将继续介绍 `kube-apiserver` 的 `Adimission` 准入流程。
