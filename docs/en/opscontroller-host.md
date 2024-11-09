### **Ops-controller-manager Host Object**

The `Host` object in the Ops Controller is used to define the configuration for individual hosts that will be managed by Ops. You can create and manage `Host` objects using `opscli` commands or YAML files.

#### **Create Host Using `opscli` Command**

To create a `Host` object directly using the `create` sub-command:

```bash
/usr/local/bin/opscli create host --name dev1 -i 1.1.1.1 --port 2222 --username root --password xxx --namespace ops-system
```

This command creates a host object with the following details:

- **`name`**: The name of the host (e.g., `dev1`).
- **`address`**: The IP address of the host (e.g., `1.1.1.1`).
- **`port`**: The SSH port for accessing the host (e.g., `2222`).
- **`username`**: The SSH username (e.g., `root`).
- **`password`**: The password for SSH access.
- **`namespace`**: The Kubernetes namespace to which the host object belongs (e.g., `ops-system`).

#### **Create Host Using YAML File**

Alternatively, you can define the `Host` object in a YAML file and apply it using `kubectl`. Here is an example YAML definition:

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Host
metadata:
  name: dev1
  namespace: ops-system
spec:
  address: 1.1.1.1
  port: 2222
  privatekey: base64 encoded private key
  username: root
  privatekeypath: ~/.ssh/id_rsa
  timeoutseconds: 10
```

In this YAML file:

- **`address`**: The IP address of the host.
- **`port`**: The SSH port number for the host.
- **`privatekey`**: The base64-encoded private SSH key for authentication.
- **`privatekeypath`**: The path to the private SSH key file.
- **`username`**: The username for SSH access.
- **`timeoutseconds`**: The timeout value (in seconds) for SSH connections.

You can apply this file using the following command:

```bash
kubectl apply -f host.yaml
```

#### **View Host Object Status**

To view the status of the `Host` object, use the following command:

```bash
kubectl get hosts dev1 -n ops-system
```

This will return information about the `Host` object, including:

- **`NAME`**: The name of the host.
- **`HOSTNAME`**: The hostname of the machine (e.g., `node1`).
- **`ADDRESS`**: The IP address of the host.
- **`DISTRIBUTION`**: The OS distribution (e.g., `centos`).
- **`ARCH`**: The architecture of the host (e.g., `x86_64`).
- **`CPU`**: The number of CPUs on the host.
- **`MEM`**: The amount of memory on the host (e.g., `7.8G`).
- **`DISK`**: The amount of disk space on the host (e.g., `52G`).
- **`HEARTTIME`**: The last time the host was checked.
- **`HEARTSTATUS`**: The status of the heartbeats (e.g., `successed`).

Example output:

```bash
NAME   HOSTNAME   ADDRESS       DISTRIBUTION   ARCH     CPU   MEM    DISK   HEARTTIME   HEARTSTATUS
dev1   node1      1.1.1.1       centos         x86_64   4     7.8G   52G    54s         successed
```
