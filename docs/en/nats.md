## NATS

### Purpose

Ops uses the NATS component to export related events, primarily of two types:

- The status of CRDs, including host and cluster states, as well as `TaskRun` and `PipelineRun` statuses.  
- Status information reported during scheduled alert inspections.  

Below is a guide for installing and configuring the NATS component. This setup follows a model of one primary cluster and multiple edge clusters. Edge clusters forward events to the primary cluster, where they are processed centrally.


### Add the Helm Repository

- Add the repository:

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- View configurable fields:

```bash
helm show values nats/nats
```


### Deploy the Primary Cluster

- Set basic information for NATS:

```bash
export adminpassword=adminpassword
export leafuser=leafuser
export leafpassword=leafpassword
```

- Generate `nats-values.yaml`:

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

This installs a core NATS server without persistence. To enable persistence, activate NATS JetStream and configure storage.

- Install NATS:

```bash
helm install nats nats/nats --version 1.2.4 -f nats-values.yaml -n ops-system
```

- Expose NATS service ports:

```bash
kubectl patch svc nats -p '{"spec":{"type":"NodePort","ports":[{"port":4222,"nodePort":32223,"targetPort":"nats"},{"port":7422,"nodePort":32222,"targetPort":"leafnodes"}]}}' -n ops-system
```

- Check the load:

```bash
kubectl -n ops-system get pod,svc | grep nats

pod/nats-0                         2/2     Running   0             15h
pod/nats-1                         2/2     Running   0             15h
pod/nats-2                         2/2     Running   0             15h
pod/nats-box-6bb86df889-xcr6x      1/1     Running   0             15h
service/nats            NodePort    10.100.109.24    <none>        4222:32223/TCP,7422:32222/TCP         15h
service/nats-headless   ClusterIP   None             <none>        4222/TCP,7422/TCP,6222/TCP,8222/TCP   15h
```


### Deploy Edge Nodes

- Add the repository:

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

- Set the NATS information for the primary cluster:

```bash
export nats_master=leafuser:leafpassword@x.x.x.x:32222
```

- Generate `nats-values.yaml`:

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

- Install NATS:

```bash
helm install nats nats/nats --version 1.2.4 -f nats-values.yaml -n ops-system
```


### Common NATS Commands

- Test NATS:

```bash
kubectl -n ops-system exec -it deployment/nats-box -- sh
```

- Subscribe to messages:

```bash
nats sub mysub
```

- Publish messages:

```bash
nats pub mysub "mymessage mycontent"
```

- View cluster information:

```bash
export adminpassword=adminpassword
nats server list --user=admin --password=${adminpassword}
```

- Perform a stress test:

```bash
nats bench benchsubject --pub 1 --sub 10
```


### References

- [JetStream Configuration](https://docs.nats.io/running-a-nats-service/configuration#jetstream)  
- [LeafNode Configuration](https://docs.nats.io/running-a-nats-service/configuration/leafnodes/leafnode_conf)  
- [Gateway Configuration](https://docs.nats.io/running-a-nats-service/configuration/gateways/gateway#gateway-configuration-block)  