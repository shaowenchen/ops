## Nats

### 用途

Ops 通过 Nats 组件，导出相关的事件，主要有两类:

- CRD 的状态，包括主机、集群的状态，TaskRun、PipelineRun 的状态
- alert 定时巡检上报的状态信息

下面提供 Nats 组件的安装与配置。这里采用的是，一个主集群，若干边缘集群的方式，边缘集群会将事件转发到主集群，在主集群统一进行处理。

### 添加 Helm Repo

- 添加仓库

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- 查看可配置的字段

```bash
helm show values nats/nats
```

### 部署主集群

- 设置 Nats 的基本信息

```bash
export adminpassword=adminpassword
export leafuser=leafuser
export leafpassword=leafpassword
```

- 生成 nats-values.yaml

```bash
cat <<EOF > nats-values.yaml
config:
  cluster:
    enabled: true
  leafnodes:
    enabled: true
    merge:
      authorization:
        user: ${leafuser}
        password: ${leafpassword}
  merge:
    accounts:
      SYS:
        users:
          - user: admin
            password: ${adminpassword}
    system_account: SYS
container:
  image:
    repository: nats
    tag: 2.10.20-alpine
natsBox:
  container:
    image:
      repository: natsio/nats-box
      tag: 0.14.5
reloader:
  enabled: true
  image:
    repository: natsio/nats-server-config-reloader
    tag: 0.15.1
EOF
```

这样安装的 Nats 只是安装了 core-nats 没有持久化，如果需要持久化，需要开启 nats-jetstream，但需要 PV 存储。

- 安装 nats

```bash
helm install nats nats/nats  --version 1.2.4  -f nats-values.yaml -n ops-system
```

- 暴露 Nats 服务端口

```bash
kubectl patch svc nats -p '{"spec":{"type":"NodePort","ports":[{"port":4222,"nodePort":32223,"targetPort":"nats"},{"port":7422,"nodePort":32222,"targetPort":"leafnodes"}]}}' -n ops-system
```

- 查看负载

```bash
kubectl -n ops-system get pod,svc | grep nats

pod/nats-0                         2/2     Running   0             15h
pod/nats-1                         2/2     Running   0             15h
pod/nats-2                         2/2     Running   0             15h
pod/nats-box-6bb86df889-xcr6x      1/1     Running   0             15h
service/nats            NodePort    10.100.109.24    <none>        4222:32223/TCP,7422:32222/TCP         15h
service/nats-headless   ClusterIP   None             <none>        4222/TCP,7422/TCP,6222/TCP,8222/TCP   15h
```

### 部署边缘节点

- 添加仓库

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- 设置主集群的 nats 信息

```bash
export nats_master=leafuser:leafpassword@x.x.x.x:32222
```

- 生成 nats-values.yaml

```bash
cat <<EOF > nats-values.yaml
config:
  leafnodes:
    enabled: true
    merge: {"remotes": [{"urls": ["nats://${nats_master}"]}]}
container:
  image:
    repository: nats
    tag: 2.10.20-alpine
natsBox:
  container:
    image:
      repository: natsio/nats-box
      tag: 0.14.5
reloader:
  enabled: true
  image:
    repository: natsio/nats-server-config-reloader
    tag: 0.15.1
EOF
```

- 安装 nats

```bash
helm install nats nats/nats  --version 1.2.4  -f nats-values.yaml -n ops-system
```

### Nats 常用命令

- 测试 Nats

```bash
kubectl -n ops-system exec -it deployment/nats-box -- sh
```

- 订阅消息

```bash
nats sub mysub
```

- 发布消息

```bash
nats pub mysub "mymessage mycontent"
```

- 查看集群信息

```bash
export adminpassword=adminpassword
nats server list --user=admin --password=${adminpassword}
```

- 压力测试

```bash
nats bench benchsubject --pub 1 --sub 10
```

### 参考

https://docs.nats.io/running-a-nats-service/configuration#jetstream
https://docs.nats.io/running-a-nats-service/configuration/leafnodes/leafnode_conf
https://docs.nats.io/running-a-nats-service/configuration/gateways/gateway#gateway-configuration-block

