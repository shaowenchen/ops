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
-i ~/.kube/config --all
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
/usr/local/bin/opscli task -f ~/.ops/task/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i 1.1.1.1 --port 2222 --username root
```

如果需要清理加上 `--clear` 参数即可。

- 集群全部节点

```bash
/usr/local/bin/opscli task -f ~/.ops/task/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i ~/.kube/config --all
```

- 集群指定节点

```bash
/usr/local/bin/opscli task -f ~/.ops/task/set-hosts.yaml --ip 1.2.3.4 --domain test.com --i ~/.kube/config --nodename node1
```

### 应用安装

- 安装 Istio

```bash
/usr/local/bin/opscli task -f ~/.ops/task/app-istio.yaml --version 1.13.7 --kubeconfig /etc/kubernetes/admin.conf
```

--version 默认值为 1.13.7，--kubeconfig 默认值为 /etc/kubernetes/admin.conf。

- 卸载 Istio

```bash
/usr/local/bin/opscli task -f ~/.ops/task/app-istio.yaml --version 1.13.7 --kubeconfig /etc/kubernetes/admin.conf --action delete
```