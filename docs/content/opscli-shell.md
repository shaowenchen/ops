## opscli shell command

### 主机清单

`-i` 参数指定

- 直接 ip

`-i 1.1.1.1`

- 批量 ip

`-i hosts.txt`

```bash
cat hosts.txt

1.1.1.1
2.2.2.2
```

opscli 会从每行中正则匹配 ip 地址，作为目标地址。

### 批量安装

```bash
opscli shell --content "curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

### 批量升级

```bash
opscli shell --content "sudo /usr/local/bin/opscli upgrade" -i hosts.txt
```

### 查找并替换镜像

```bash
opscli shell --content "sudo /usr/local/bin/opscli task -f .ops/task/list-podimage.yaml --namespace all" -i hosts.txt
```

```bash
opscli shell --content "sudo kubectl -n kube-system set image deployment/metrics-server metrics-server=hubimage/metrics-server:v0.5.0" -i hosts.txt
```

```bash
opscli shell --content "sudo kubectl -n kube-system set image deployment/metrics-server metrics-server=hubimage/metrics-server:v0.6.1" -i hosts-B.txt
```

```bash
opscli shell --content "sudo kubectl -n kube-system set image deployment/prom-k8s-kube-state-metrics kube-state-metrics=hubimage/kube-state-metrics:v2.2.4" -i hosts.txt
```
