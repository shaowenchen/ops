### opscli Task Command Guide

#### 1. **Specify Target Hosts**

- **Single Host**

```bash
-i 1.1.1.1
```

You can specify the username and password for the host with `--username` and `--password`.

- **Batch Hosts**

To specify multiple hosts via a file:

```bash
-i hosts.txt
```

Example content of `hosts.txt`:

```bash
1.1.1.1
2.2.2.2
```

- **All Nodes in a Cluster**

```bash
-i ~/.kube/config --nodename all
```

- **Specific Node in a Cluster**

```bash
-i ~/.kube/config --nodename node1
```

Where `node1` is the node name.

#### 2. **Update `/etc/hosts`**

- **On a Single Host**

To remotely update the `/etc/hosts` file on host `1.1.1.1`:

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i 1.1.1.1 --port 2222 --username root
```

To clear the `/etc/hosts` entry, add `--clear`:

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i 1.1.1.1 --clear
```

- **On All Nodes in a Cluster**

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i ~/.kube/config --nodename all
```

- **On a Specific Node in a Cluster**

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i ~/.kube/config --nodename node1
```

#### 3. **Install Applications**

- **Install Istio**

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/app-istio.yaml --version 1.13.7 --kubeconfig /etc/kubernetes/admin.conf
```

The default version is `1.13.7` and the default `kubeconfig` is `/etc/kubernetes/admin.conf`.

- **Uninstall Istio**

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/app-istio.yaml --version 1.13.7 --kubeconfig /etc/kubernetes/admin.conf --action delete
```

#### 4. **Upload Files**

- **Upload to Server**

```bash
/usr/local/bin/opscli task -f tasks/file-upload.yaml --fileapi https://gh-uploadapi.chenshaowen.com/api/v1/files --localfile dockerfile
```

- **Upload to S3**

```bash
/usr/local/bin/opscli task -f tasks/file-upload.yaml --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket xxx --localfile dockerfile --remotefile s3://dockerfile
```

#### 5. **Download Files**

- **From Server**

```bash
/usr/local/bin/opscli task -f tasks/file-download.yaml --fileapi https://gh-uploadapi.chenshaowen.com/api/v1/files --aeskey <AES_KEY> --remotefile <URL> --localfile dockerfile1
```

- **From S3**

```bash
/usr/local/bin/opscli task -f tasks/file-download.yaml --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket xxx --localfile dockerfile2 --remotefile s3://dockerfile
```
