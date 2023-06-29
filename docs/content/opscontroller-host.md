## Ops-controller-manager host 对象

### 直接使用 `create` 子命令创建

```bash
/usr/local/bin/opscli create host --name dev1 -i 1.1.1.1 --port 2222 --username root --password xxx --namespace ops-system
```

### 使用 yaml 文件创建

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

### 查看对象

```bash
kubectl get hosts dev1 -n ops-system

NAME   HOSTNAME   ADDRESS       DISTRIBUTION   ARCH     CPU   MEM    DISK   HEARTTIME   HEARTSTATUS
dev1   node1      1.1.1.1       centos         x86_64   4     7.8G   52G    54s         successed
```
