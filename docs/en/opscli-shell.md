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
