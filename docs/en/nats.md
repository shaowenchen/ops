## NATS

### Purpose

Ops uses the NATS component to export relevant events, primarily of two types:

- The status of CRDs, including the status of hosts, clusters, TaskRun, and PipelineRun.
- Status information reported by scheduled inspections from alerts.

Below is a guide for installing and configuring the NATS component. This setup follows a model with one primary cluster and multiple edge clusters. The edge clusters forward events to the primary cluster for unified processing.

### Adding the Helm Repo

- Add the repository:

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- View configurable fields:

```bash
helm show values nats/nats
```

### Deploying the Primary Cluster

- Set basic NATS credentials:

```bash
export adminpassword=mypassword
export apppassword=mypassword
```

- set nats server name

```bash
export natsservername=need-to-be-unique
```

- Generate `nats-values.yaml`:

```bash
cat <<EOF > nats-values.yaml
config:
  leafnodes:
    enabled: true
    merge:
      remotes:
        - urls:
          - nats://admin:${adminpassword}@${natsendpoint}
          account: SYS
        - urls:
          - nats://app:${apppassword}@${natsendpoint}
          account: APP
  merge:
    server_name: ${natsservername}
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
    repository: registry.cn-beijing.aliyuncs.com/opshub/nats
    tag: 2.10.20-alpine
natsBox:
  container:
    image:
      repository: registry.cn-beijing.aliyuncs.com/opshub/natsio-nats-box
      tag: 0.14.5
reloader:
  enabled: true
  image:
    repository: registry.cn-beijing.aliyuncs.com/opshub/natsio-nats-server-config-reloader
    tag: 0.15.1
EOF
```

The data is persisted in memory. To store it on disk, enable the `fileStore` configuration.

- Install NATS:

```bash
helm -n ops-system install nats nats/nats  --version 1.2.4  -f nats-values.yaml
```

- Uninstall NATS:

```bash
helm -n ops-system uninstall nats
```

- Expose the NATS service ports:

```bash
kubectl patch svc nats -p '{"spec":{"type":"NodePort","ports":[{"port":4222,"nodePort":32223,"targetPort":"nats"},{"port":7422,"nodePort":32222,"targetPort":"leafnodes"}]}}' -n ops-system
```

- Check the workload:

```bash
kubectl -n ops-system get pod,svc | grep nats

pod/nats-0                         2/2     Running   0             15h
pod/nats-1                         2/2     Running   0             15h
pod/nats-2                         2/2     Running   0             15h
pod/nats-box-6bb86df889-xcr6x      1/1     Running   0             15h
service/nats            NodePort    10.100.109.24    <none>        4222:32223/TCP,7422:32222/TCP         15h
service/nats-headless   ClusterIP   None             <none>        4222/TCP,7422/TCP,6222/TCP,8222/TCP   15h
```

### Deploying Edge Clusters

- Add the repository:

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- Set the primary cluster's NATS information:

```bash
export natsendpoint=x.x.x.x:32222
```

- Generate `nats-values.yaml`:

Note that the `server_name` must be unique for each cluster; otherwise, duplicate connection issues will arise.

```bash
cat <<EOF > nats-values.yaml
config:
  leafnodes:
    enabled: true
    merge:
      remotes:
        - urls:
          - nats://admin:${adminpassword}@${natsendpoint}
          account: SYS
        - urls:
          - nats://app:${apppassword}@${natsendpoint}
          account: APP
  merge:
    server_name: need-to-be-unique
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

- Install NATS:

```bash
helm install nats nats/nats  --version 1.2.4  -f nats-values.yaml -n ops-system
```

### Common NATS Commands

- Test NATS:

```bash
kubectl -n ops-system exec -it deployment/nats-box -- sh
```

- Subscribe to messages:

```bash
nats --user=app --password=${apppassword} sub "ops.>"
```

- Publish messages:

```bash
nats --user=app --password=${apppassword} pub ops.test "mymessage mycontent"
```

- Create a stream to persist messages:

```bash
nats --user=app --password=${apppassword} stream add ops --subjects "ops.>" --ack --max-msgs=-1 --max-bytes=-1 --max-age=168h --storage file --retention limits --max-msg-size=-1 --discard=old --replicas 1 --dupe-window=2m
```

For production environments, it is recommended to use file storage and set replicas to 3.

- View stream events:

```bash
nats --user=app --password=${apppassword} stream view ops
```

- View stream configuration:

```bash
nats --user=app --password=${apppassword} stream info ops
```

- View cluster information:

```bash
nats --user=admin --password=${adminpassword} server report jetstream
```

This command displays information about the primary cluster, edge clusters, and their connections.

- View the subjects of a stream:

```bash
nats --user=app --password=${adminpassword} stream subjects ops
```

- Perform a benchmark:

```bash
nats --user=app --password=${apppassword} bench benchsubject --pub 1 --sub 10
```

### References

- [JetStream Configuration](https://docs.nats.io/running-a-nats-service/configuration#jetstream)
- [LeafNode Configuration](https://docs.nats.io/running-a-nats-service/configuration/leafnodes/leafnode_conf)
- [Gateway Configuration](https://docs.nats.io/running-a-nats-service/configuration/gateways/gateway#gateway-configuration-block)
