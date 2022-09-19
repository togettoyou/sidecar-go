# sidecar-go

通过 Admission Webhook 拦截并修改 Kubernetes API Server 收到的请求，自动向符合条件的 Pod 注入 sidecar 容器。

### 安装

```shell
$ kubectl apply -f https://raw.githubusercontent.com/togettoyou/sidecar-go/main/config/deploy.yaml
```

### 快速入门

1、创建 SidecarGo

```yaml
# sidecargo-sample.yaml
apiVersion: apps.togettoyou.com/v1alpha1
kind: SidecarGo
metadata:
  name: sidecargo-sample
spec:
  selector:
    matchLabels: # 通过标签匹配 Pod
      app: nginx
  initContainers: # 需要注入的 init 容器列表
    - name: init-container
      image: busybox:1.28.4
      command: [ "/bin/sh", "-c", "sleep 5 && echo 'init container success'" ]
  containers: # 需要注入的 sidecar 容器列表
    - name: sidecar
      image: busybox:1.28.4
      command: [ "sleep", "3600" ]
      volumeMounts:
        - name: log-volume
          mountPath: /var/log
  volumes: # 需要注入的 volumes 列表
    - name: log-volume
      emptyDir: { }
```

```shell
$ kubectl apply -f sidecargo-sample.yaml
sidecargo.apps.togettoyou.com/sidecargo-sample created
$ kubectl get sidecargo
NAME               AGE
sidecargo-sample   17s
```

2、创建 Pod

```yaml
# pod-nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  containers:
    - name: nginx
      image: nginx:latest
      ports:
        - containerPort: 80
```

```shell
$ kubectl apply -f pod-nginx.yaml
pod/nginx created
```

3、验证

```shell
$ kubectl get pod
NAME    READY   STATUS            RESTARTS   AGE
nginx   0/2     PodInitializing   0          14s
$ kubectl get pod
NAME    READY   STATUS    RESTARTS   AGE
nginx   2/2     Running   0          16s
```

### 卸载

```shell
$ kubectl delete -f https://raw.githubusercontent.com/togettoyou/sidecar-go/main/config/deploy.yaml
```