## Ops-controller-manager cluster 对象

### 直接使用 `create` 子命令创建

```bash
/usr/local/bin/opscli create cluster -i  ~/.kube/config --name dev1 --namespace ops-system
```

### 使用 yaml 文件创建

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Cluster
metadata:
  name: dev1
  namespace: ops-system
spec:
  config: base64 encoded kubeadm config
  server: https://1.1.1.1:6443
```

### 查看对象

```bash
kubectl get cluster dev1 -n ops-system

NAME   SERVER                     VERSION   NODE   RUNNING   TOTALPOD   CERTDAYS   STATUS
dev1   https://1.1.1.1:6443       v1.21.0   1      15        16         114        successed
```
