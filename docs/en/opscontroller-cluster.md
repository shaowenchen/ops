### **Ops-controller-manager Cluster Object**

The `Cluster` object in the Ops Controller can be created and managed using `opscli` commands or YAML files.

#### **Create Cluster Using `opscli` Command**

To create a cluster directly using the `create` sub-command:

```bash
/usr/local/bin/opscli create cluster -i ~/.kube/config --name dev1 --namespace ops-system
```

This command creates a cluster named `dev1` in the `ops-system` namespace, using the default kubeconfig file located at `~/.kube/config`.

#### **Create Cluster Using YAML File**

Alternatively, you can define the `Cluster` object in a YAML file and apply it using `kubectl`. Here is an example YAML definition:

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Cluster
metadata:
  name: dev1
  namespace: ops-system
spec:
  config: base64 encoded kubeadm config
  server: https://1.1.1.1:6443
```

In this YAML file:

- **`config`**: Should contain the base64-encoded `kubeadm` configuration.
- **`server`**: The address of the Kubernetes API server.

You can apply this file using the following command:

```bash
kubectl apply -f cluster.yaml
```

#### **View Cluster Object Status**

To view the status of the `Cluster` object, use the following command:

```bash
kubectl get cluster dev1 -n ops-system
```

This will return information about the `Cluster` object, including:

- **`NAME`**: The name of the cluster.
- **`SERVER`**: The address of the Kubernetes API server.
- **`VERSION`**: The Kubernetes version.
- **`NODE`**: The number of nodes.
- **`RUNNING`**: The number of running nodes.
- **`TOTALPOD`**: The total number of pods in the cluster.
- **`CERTDAYS`**: The remaining days of the cluster's certificate validity.
- **`STATUS`**: The current status of the cluster (e.g., `successed`).

Example output:

```bash
NAME   SERVER                     VERSION   NODE   RUNNING   TOTALPOD   CERTDAYS   STATUS
dev1   https://1.1.1.1:6443       v1.21.0   1      15        16         114        successed
```
