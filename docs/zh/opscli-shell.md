## opscli shell command

### 指定操作目标清单

- 指定主机

`-i 1.1.1.1`

通过 `--username` 指定用户名，`--password` 指定密码。

- 批量主机

通过文件指定:

`-i hosts.txt`

```bash
cat hosts.txt

1.1.1.1
2.2.2.2
```

opscli 会从每行中正则匹配 ip 地址，作为目标地址。

通过逗号分割指定:

`-i 1.1.1.1,2.2.2.2`

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

### 查看集群镜像

- 单机

```bash
/usr/local/bin/opscli task -f ~/.ops/tasks/list-podimage.yaml --namespace all
```

### 集群批量操作

- 全部节点

```bash
opscli shell --content "uname -a" --nodename all
```

- 指定节点

```bash
opscli shell --content "uname -a" --nodename node1
```

- 指定 kubeconfig

默认 kubeconfig 为 `~/.kube/config`，可以通过 `-i` 参数指定。

```bash
opscli shell -i  ~/Documents/opscli/prod --content "uname -a" --nodename node1
```
