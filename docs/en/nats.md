## Nats

### Purpose

Ops uses the Nats component to export relevant events, primarily in two categories:

- CRD status, including the status of hosts, clusters, TaskRun, and PipelineRun.
- Alert status information reported by scheduled inspections.

Below is the installation and configuration for the Nats component. We use one main cluster and several edge clusters, where the edge clusters forward events to the main cluster for centralized processing.

### Add Helm Repo

- Add repository

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- View configurable fields

```bash
helm show values nats/nats
```

### Deploy the Main Cluster

- Set basic Nats information

```bash
export adminpassword=adminpassword
export leafuser=leafuser
export leafpassword=leafpassword
export apppassword=apppassword
```

- Generate `nats-values.yaml`

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

This Nats installation only installs the core Nats without persistence. To enable persistence, Jetstream must be enabled, and storage should be configured.

- Install Nats

```bash
helm install nats nats/nats  --version 1.2.4  -f nats-values.yaml -n ops-system
```

- Expose Nats service port

```bash
kubectl patch svc nats -p '{"spec":{"type":"NodePort","ports":[{"port":4222,"nodePort":32223,"targetPort":"nats"},{"port":7422,"nodePort":32222,"targetPort":"leafnodes"}]}}' -n ops-system
```

- View load status

```bash
kubectl -n ops-system get pod,svc | grep nats

pod/nats-0                         2/2     Running   0             15h
pod/nats-1                         2/2     Running   0             15h
pod/nats-2                         2/2     Running   0             15h
pod/nats-box-6bb86df889-xcr6x      1/1     Running   0             15h
service/nats            NodePort    10.100.109.24    <none>        4222:32223/TCP,7422:32222/TCP         15h
service/nats-headless   ClusterIP   None             <none>        4222/TCP,7422/TCP,6222/TCP,8222/TCP   15h
```

### Deploy Edge Node

- Add repository

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- Set Nats information for the main cluster

```bash
export nats_master=leafuser:leafpassword@x.x.x.x:32222
```

- Generate `nats-values.yaml`

Note that the `server_name` for different clusters must not be the same, as this would cause duplicate connection issues.

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

- Install Nats

```bash
helm install nats nats/nats  --version 1.2.4  -f nats-values.yaml -n ops-system
```

### Common Nats Commands

- Test Nats

```bash
kubectl -n ops-system exec -it deployment/nats-box -- sh
```

- Subscribe to a message

```bash
nats sub ops.* --user=app --password=${apppassword}
```

- Publish a message

```bash
nats pub ops.* "mymessage mycontent" --user=app --password=${apppassword}
```

- Create a stream to persist messages

```bash
nats stream add ops --subjects "ops.*" --ack --max-msgs=-1 --max-bytes=-1 --max-age=1y --storage file --retention limits --max-msg-size=-1 --discard=old --replicas 3 --dupe-window=2m --user=app --password=${apppassword}
```

- View stream information

```bash
nats stream view ops --user=app --password=${apppassword}
```

- View stream configuration

```bash
nats stream info ops --user=app --password=${apppassword}
```

- View cluster information

```bash
nats server list --user=admin --password=${adminpassword}
```

- Perform a stress test

```bash
nats bench benchsubject --pub 1 --sub 10 --user=app --password=${apppassword}
```

### References

- [NATS JetStream Configuration](https://docs.nats.io/running-a-nats-service/configuration#jetstream)
- [NATS Leafnode Configuration](https://docs.nats.io/running-a-nats-service/configuration/leafnodes/leafnode_conf)
- [NATS Gateway Configuration](https://docs.nats.io/running-a-nats-service/configuration/gateways/gateway#gateway-configuration-block)
