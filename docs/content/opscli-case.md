## opscli 使用案例

### 在 kubectl pod 中测试指定节点的磁盘 IO 性能

- 安装 opscli for alpine

```bash
sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
apk add curl
curl -sfL https://mirror.ghproxy.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh |VERSION=latest sh -
```

- 在节点安装 fio

```bash
opscli shell --content "apt-get install fio -y" --nodename node1
```

- 在节点上测试磁盘 IO 性能

```bash
opscli task -f ~/.ops/tasks/get-diskio-byfio.yaml --size 1g --filename=/tmp/testfile --nodename node1
```

其中 size 为测试文件大小，filename 为测试文件路径，nodename 为测试节点名称。

```
(1/8) Rand_Read_Testing

read: IOPS=105k, BW=410MiB/s (430MB/s)(1024MiB/2498msec) -> 4k 随机读 410 MiB/s

(2/8) Rand_Write_Testing

write: IOPS=55.9k, BW=218MiB/s (229MB/s)(1024MiB/4688msec) -> 4k 随机写 218 MiB/s

(3/8) Sequ_Read_Testing

read: IOPS=51.8k, BW=6481MiB/s (6796MB/s)(1024MiB/158msec) -> 128k 顺序读 6481 MiB/s

(4/8) Sequ_Write_Testing

write: IOPS=30.7k, BW=3835MiB/s (4022MB/s)(1024MiB/267msec) -> 128k 顺序写 3835 MiB/s

(5/8) Rand_Read_IOPS_Testing

read: IOPS=80.4k, BW=314MiB/s (329MB/s)(1024MiB/3261msec) -> 4k 下读 IOPS 为 80.4k

(6/8) Rand_Write_IOPS_Testing

write: IOPS=83.4k, BW=326MiB/s (342MB/s)(1024MiB/3143msec) -> 4k 下写 IOPS 为 83.4k

(7/8) Rand_Read_Latency_Testing

lat (usec): min=34, max=457722, avg=57.78, stdev=1630.32 -> 4k 读延时为 57.78 us

(8/8) Rand_Write_Latency_Testing

lat (usec): min=35, max=664838, avg=385.12, stdev=5335.64 -> 4k 写延时为 385.12 us
```

### 给集群 GPU 主机配置巡检

- 在全部 master 节点上安装 Opscli

```bash
opscli task -f ~/.ops/tasks/install-opscli.yaml -i master-ips.txt
```

- 在能 ssh 全部节点的机器上，创建访问主机的 ssh 密钥

```bash
kubectl -n ops-system create secret generic host-secret --from-file=privatekey=/root/.ssh/id_rsa
```

- 自动发现集群主机

```bash
kubectl apply -f ~/.ops/tasks/auto-create-host.yaml
```

- 自动给 host 对象打上标签

```bash
kubectl apply -f ~/.ops/tasks/alert-label-gpu.yaml
```

- 配置巡检任务

```bash
kubectl apply -f ~/.ops/tasks/alert-gpu-drop.yaml
```

```bash
kubectl apply -f ~/.ops/tasks/alert-gpu-zombie.yaml
```

- 自动清理 Ops 运行 Task 创建的临时 Pod 对象

```bash
kubectl apply -f ~/.ops/tasks/clear-opstaskpod.yaml
```
