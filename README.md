# sidecar-go

> 一个自动向 Pod 注入 sidecar 容器的 k8s 控制器

原理：通过 Admission Webhook 拦截并修改 Kubernetes API Server 收到的请求

框架：[operator-sdk](https://github.com/operator-framework/operator-sdk)

```shell
$ operator-sdk init --domain togettoyou.com --repo github.com/togettoyou/sidecar-go
$ operator-sdk create api --group apps --version v1alpha1 --kind SidecarGo --resource --controller
```