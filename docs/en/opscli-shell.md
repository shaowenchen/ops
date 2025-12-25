### opscli Shell Command Guide

#### 1. **Specify Target Hosts**

- **Single Host**

Use the `-i` flag to specify a single host IP.

```bash
-i 1.1.1.1
```

You can also specify a username and password with the `--username` and `--password` flags:

```bash
--username <your-username> --password <your-password>
```

- **Batch Hosts**

To specify multiple hosts, you can use a file or comma-separated IP addresses.

- **From a File (`hosts.txt`)**

```bash
-i hosts.txt
```

Example content of `hosts.txt`:

```bash
1.1.1.1
2.2.2.2
```

- **Comma-separated IPs**

```bash
-i 1.1.1.1,2.2.2.2
```

- **All Nodes in a Cluster**

```bash
-i ~/.kube/config --nodename all
```

By default, `-i` points to `~/.kube/config`.

- **Specific Node in a Cluster**

```bash
-i ~/.kube/config --nodename node1
```

Where `node1` is the node name.

#### 2. **View Cluster Images**

- **For Single Machine**

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/list-podimage.yaml --namespace all
```

#### 3. **Cluster Bulk Operations**

- **All Nodes**

To run a command on all nodes:

```bash
opscli shell --content "uname -a" --nodename all
```

- **Specific Node**

To run a command on a specific node:

```bash
opscli shell --content "uname -a" --nodename node1
```

- **Specify kubeconfig**

To specify a custom `kubeconfig`, use the `-i` flag:

```bash
opscli shell -i ~/Documents/opscli/prod --content "uname -a" --nodename node1
```

#### 4. **Mount Host Paths to Container**

The `--mount` flag allows you to mount host paths into the container. This is useful for accessing host files, directories, or socket files (like `docker.sock`) from within the container.

- **Single Mount**

```bash
opscli shell --content "ls /data" --mount /opt/data:/data
```

- **Multiple Mounts**

You can specify multiple mounts by using `--mount` multiple times:

```bash
opscli shell --content "ls /data /logs" \
    --mount /opt/data:/data \
    --mount /opt/logs:/logs
```

- **Mount Docker Socket**

To mount the Docker socket and use Docker commands inside the container:

```bash
opscli shell --content "docker ps" \
    --mount /var/run/docker.sock:/var/run/docker.sock
```

- **Mount Format**

The mount format is: `hostPath:mountPath`
- `hostPath`: absolute path on the host (required)
- `mountPath`: absolute path in the container (required)

**Note**: If you need to access the host root filesystem, you can mount it explicitly:
```bash
opscli shell --content "ls /host" --mount /:/host
```
