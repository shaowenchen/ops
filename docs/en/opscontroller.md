### **Ops-controller-manager Overview**

`ops-controller-manager` is a Kubernetes Operator that provides three core objects: `Host`, `Cluster`, and `Task`.

#### **Objects in ops-controller-manager**

- **Host**: Describes a host machine, including hostname, IP address, SSH username, password, private key, etc.
- **Cluster**: Describes a cluster, including details such as cluster name, number of hosts, number of pods, load, CPU, memory, etc.
- **Task**: Describes a task, which can be one-time or scheduled (cron) tasks.

### **Installation**

#### **Install Helm**

If you haven't installed Helm yet, you can install it using the following command:

```bash
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

#### **Add Helm Repository**

Add the `ops` Helm repository to your Helm configuration:

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
helm repo update
```

#### **Install ops-controller-manager**

To install the `ops-controller-manager` using Helm:

```bash
helm install myops ops/ops --version 2.0.0 --namespace ops-system --create-namespace
```

This command installs the `ops-controller-manager` in the `ops-system` namespace and creates the namespace if it doesn't exist.

#### **Verify Installation**

After the installation, check the status of the pods to ensure everything is running:

```bash
kubectl get pods -n ops-system
```

#### **Uninstall ops-controller-manager**

To uninstall the `ops-controller-manager`:

```bash
helm -n ops-system uninstall myops
```

### **Namespace Configuration**

By default, `ops-controller-manager` will only process CRD resources within the `ops-system` namespace.

To change this behavior, you can modify the `ACTIVE_NAMESPACE` environment variable in the configuration. If left empty, it will process resources from all namespaces.

Make sure to update the namespace if you need `ops-controller-manager` to work with resources from a different namespace.
