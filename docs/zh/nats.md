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
export apppassword=apppassword
```

- 生成 nats-values.yaml

```bash
cat <<EOF > nats-values.yaml
config:
  jetstream:
    enabled: false
    fileStore:
      enabled: true
      dir: /data
    pvc:
      enabled: true
      storageClassName: my-sc-client
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
      APP:
        users:
          - user: app
            password: ${apppassword}
        jetstream: true
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

这样安装的 Nats 只是安装了 core-nats 没有持久化，如果需要持久化，需要开启 nats-jetstream，但需要配置存储。

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

需要注意的是，不同集群的 `server_name` 不能相同，否则会有重复连接的问题。

```bash
cat <<EOF > nats-values.yaml
config:
  leafnodes:
    enabled: true
    merge: {"remotes": [{"urls": ["nats://${nats_master}"]}]}
  merge:
    server_name: nats-cluster-1
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
nats sub ops.* --user=app --password=${apppassword}
```

- 发布消息

```bash
nats pub ops.* "mymessage mycontent" --user=app --password=${apppassword}
```

- 创建 stream 持久化消息

```bash
nats stream add ops --subjects "ops.*" --ack --max-msgs=-1 --max-bytes=-1 --max-age=1y --storage file --retention limits --max-msg-size=-1 --discard=old --replicas 3 --dupe-window=2m --user=app --password=${apppassword}
```

- 查看 stream 信息

```bash
nats stream view ops --user=app --password=${apppassword}
```

- 查看 stream 配置

```bash
nats stream info ops --user=app --password=${apppassword}
```

- 查看集群信息

```bash
nats server list --user=admin --password=${adminpassword}
```

- 压力测试

```bash
nats bench benchsubject --pub 1 --sub 10 --user=app --password=${apppassword}
```

### 参考

https://docs.nats.io/running-a-nats-service/configuration#jetstream
https://docs.nats.io/running-a-nats-service/configuration/leafnodes/leafnode_conf
https://docs.nats.io/running-a-nats-service/configuration/gateways/gateway#gateway-configuration-block
