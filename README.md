# sidecar-go

> 一个自动向 Pod 注入 sidecar 容器的 k8s 控制器

原理：通过 Admission Webhook 拦截并修改 Kubernetes API Server 收到的请求

框架：[operator-sdk](https://github.com/operator-framework/operator-sdk)

项目初始化：

```shell
$ operator-sdk init --domain togettoyou.com --repo github.com/togettoyou/sidecar-go
```

`SidecarGo` 类型及其 Controller ：

```shell
$ operator-sdk create api --group apps --version v1alpha1 --kind SidecarGo --resource --controller
```

`Pod` 类型及其 Webhook ：

```shell
$ operator-sdk create api --group core --version v1 --kind Pod --resource=false --controller=false
$ operator-sdk create webhook --group core --version v1 --kind Pod --defaulting --programmatic-validation
```