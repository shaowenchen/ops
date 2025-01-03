## opscli Usage Examples

### Test Disk IO Performance on a Specific Node in kubectl Pod

- Install opscli for alpine

```bash
sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
apk add curl
curl -sfL https://ghproxy.chenshaowen.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh |VERSION=latest sh -
```

- Install fio on the node

```bash
opscli shell --content "apt-get install fio -y" --nodename node1
```

- Test Disk IO performance on the node

```bash
opscli task -f ~/.ops/tasks/get-diskio-byfio.yaml --size 1g --filename=/tmp/testfile --nodename node1
```

Where `size` is the test file size, `filename` is the test file path, and `nodename` is the test node name.

```
(1/8) Rand_Read_Testing

read: IOPS=105k, BW=410MiB/s (430MB/s)(1024MiB/2498msec) -> 4k Random Read 410 MiB/s

(2/8) Rand_Write_Testing

write: IOPS=55.9k, BW=218MiB/s (229MB/s)(1024MiB/4688msec) -> 4k Random Write 218 MiB/s

(3/8) Sequ_Read_Testing

read: IOPS=51.8k, BW=6481MiB/s (6796MB/s)(1024MiB/158msec) -> 128k Sequential Read 6481 MiB/s

(4/8) Sequ_Write_Testing

write: IOPS=30.7k, BW=3835MiB/s (4022MB/s)(1024MiB/267msec) -> 128k Sequential Write 3835 MiB/s

(5/8) Rand_Read_IOPS_Testing

read: IOPS=80.4k, BW=314MiB/s (329MB/s)(1024MiB/3261msec) -> 4k Read IOPS 80.4k

(6/8) Rand_Write_IOPS_Testing

write: IOPS=83.4k, BW=326MiB/s (342MB/s)(1024MiB/3143msec) -> 4k Write IOPS 83.4k

(7/8) Rand_Read_Latency_Testing

lat (usec): min=34, max=457722, avg=57.78, stdev=1630.32 -> 4k Read Latency 57.78 us

(8/8) Rand_Write_Latency_Testing

lat (usec): min=35, max=664838, avg=385.12, stdev=5335.64 -> 4k Write Latency 385.12 us
```

### Configure Inspection for GPU Hosts in the Cluster

- Install Opscli on all master nodes

```bash
opscli task -f ~/.ops/tasks/install-opscli.yaml -i master-ips.txt
```

- Create an SSH key to access the hosts from a machine that can SSH into all nodes

```bash
kubectl -n ops-system create secret generic host-secret --from-file=privatekey=/root/.ssh/id_rsa
```

- Add all task templates

```bash
kubectl apply -f ~/.ops/tasks/
```

- Automatically discover hosts

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: auto-create-host
  namespace: ops-system
spec:
  crontab: 40 * * * *
  taskRef: auto-create-host
EOF
```

- Automatically label hosts

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: alert-label-gpu
  namespace: ops-system
spec:
  crontab: 40 * * * *
  taskRef: alert-label-gpu
EOF
```

- GPU Card Drop Inspection

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: alert-gpu-drop
  namespace: ops-system
spec:
  crontab: 40 * * * *
  taskRef: alert-gpu-drop
EOF
```

- GPU ECC Inspection

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: alert-gpu-ecc
  namespace: ops-system
spec:
  crontab: 40 * * * *
  taskRef: alert-gpu-ecc
EOF
```

- GPU Fabric Inspection

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: alert-gpu-fabric
  namespace: ops-system
spec:
  crontab: 40 * * * *
  taskRef: alert-gpu-fabric
EOF
```

- GPU Zombie Inspection

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: alert-gpu-zombie
  namespace: ops-system
spec:
  crontab: 40 * * * *
  taskRef: alert-gpu-zombie
EOF
```

- Disk Cleanup Scheduling

```bash
kubectl apply -f - <<EOF
apiVersion: crd.chenshaowen.com/v1
kind: TaskRun
metadata:
  name: clear-disk
  namespace: ops-system
spec:
  crontab: 0 1 * * *
  taskRef: clear-disk
EOF
```
