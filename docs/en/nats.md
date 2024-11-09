## Nats

### Purpose

Ops exports related events through the Nats component, which mainly includes two types:

- CRD statuses, including host and cluster statuses, TaskRun, PipelineRun statuses.
- Status information reported by alert scheduled inspections.

Here is the installation and configuration for the Nats component. The setup uses one main cluster and several edge clusters. The edge clusters forward events to the main cluster for unified processing.

### Add Helm Repo

- Add the repository

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- View configurable fields

```bash
helm show values nats/nats
```

### Deploy Main Cluster

- Set basic Nats information

```bash
export adminpassword=adminpassword
export leafuser=leafuser
export leafpassword=leafpassword
```

- Generate `nats-values.yaml`

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

The Nats installed this way only includes the core Nats without persistence. To enable persistence, `nats-jetstream` needs to be activated, but it requires PV storage.

- Install Nats

```bash
helm install nats nats/nats  --version 1.2.4  -f nats-values.yaml -n ops-system
```

- Expose Nats service ports

```bash
kubectl patch svc nats -p '{"spec":{"type":"NodePort","ports":[{"port":4222,"nodePort":32223,"targetPort":"nats"},{"port":7422,"nodePort":32222,"targetPort":"leafnodes"}]}}' -n ops-system
```

- View load

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

- Add the repository

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- Set the Nats information of the main cluster

```bash
export nats_master=leafuser:leafpassword@x.x.x.x:32222
```

- Generate `nats-values.yaml`

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
nats sub mysub
```

- Publish a message

```bash
nats pub mysub "mymessage mycontent"
```

- View cluster information

```bash
export adminpassword=adminpassword
nats server list --user=admin --password=${adminpassword}
```

- Stress test

```bash
nats bench benchsubject --pub 1 --sub 10
```

### References

https://docs.nats.io/running-a-nats-service/configuration#jetstream  
https://docs.nats.io/running-a-nats-service/configuration/leafnodes/leafnode_conf  
https://docs.nats.io/running-a-nats-service/configuration/gateways/gateway#gateway-configuration-block
