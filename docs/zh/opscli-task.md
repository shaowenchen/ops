## opscli task command

### `-i` 指定操作目标清单

- 指定主机

`-i 1.1.1.1`

通过 `--username` 指定用户名，`--password` 指定密码。

- 批量主机

`-i hosts.txt`

```bash
cat hosts.txt

1.1.1.1
2.2.2.2
```

opscli 会从每行中正则匹配 ip 地址，作为目标地址。

- 集群全部节点

```bash
-i ~/.kube/config --nodename all
```

`-i` 默认值为 `~/.kube/config`。

- 集群指定节点

```bash
-i ~/.kube/config --nodename node1
```

node1 为节点名称。

### 更新 `/etc/hosts`

- 主机

远程到主机 `1.1.1.1` ，更新 `/etc/hosts` 文件。

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i 1.1.1.1 --port 2222 --username root
```

如果需要清理加上 `--clear` 参数即可。

- 集群全部节点

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i ~/.kube/config --nodename all
```

- 集群指定节点

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i ~/.kube/config --nodename node1
```

### 应用安装

- 安装 Istio

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/app-istio.yaml --version 1.13.7 --kubeconfig /etc/kubernetes/admin.conf
```

--version 默认值为 1.13.7，--kubeconfig 默认值为 /etc/kubernetes/admin.conf。

- 卸载 Istio

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/app-istio.yaml --version 1.13.7 --kubeconfig /etc/kubernetes/admin.conf --action delete
```

### 上传文件

- 上传到 Server

```bash
/usr/local/bin/opscli task -f tasks/file-upload.yaml --api https://uploadapi.chenshaowen.com/api/v1/files --localfile dockerfile

> Run Task  ops-system/file-upload  on  127.0.0.1
(1/1) upload file
Please use the following command to download the file: 
opscli file --api https://uploadapi.chenshaowen.com/api/v1/files --aeskey a9f891afe71fda777b05a7063068360a914e83848d7da46d7513aee86c053f6c --direction download --remotefile https://uploadapi.chenshaowen.com/uploadbases/cdn0/raw/1721615659-dockerfile.aes
```

- 上传到 S3

```bash
/usr/local/bin/opscli task -f tasks/file-upload.yaml --ak xxx --sk xxx --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket xxx --localfile dockerfile --remotefile s3://dockerfile
```

### 下载文件

- 从 Server 下载

```bash
/usr/local/bin/opscli task -f task -f tasks/file-download.yaml --api https://uploadapi.chenshaowen.com/api/v1/files --aeskey a9f891afe71fda777b05a7063068360a914e83848d7da46d7513aee86c053f6c --remotefile https://uploadapi.chenshaowen.com/uploadbases/cdn0/raw/1721615659-dockerfile.aes --localfile dockerfile1

> Run Task  ops-system/file-download  on  127.0.0.1
(1/1) download file
success download https://uploadapi.chenshaowen.com/uploadbases/cdn0/raw/1721615659-dockerfile.aes to dockerfile1
```

- 从 S3 下载

```bash
/usr/local/bin/opscli task -f tasks/file-download.yaml --ak xxx --sk xxx  --region beijing --endpoint ks3-cn-beijing.ksyun.com --bucket xxx --localfile dockerfile2 --remotefile s3://dockerfile

> Run Task  ops-system/file-download  on  127.0.0.1
(1/1) download file
success download s3 dockerfile to dockerfile2
```
