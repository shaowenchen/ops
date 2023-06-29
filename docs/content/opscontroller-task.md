## Ops-controller-manager task 对象

### 直接使用 `create` 子命令创建

```bash
go run cmd/cli/main.go create task --name t1 --typeref host --nameref dev1  --filepath ./task/get-osstaus.yaml
```

通过 `--typeref host --nameref dev1` 指定任务执行的主机。

### 使用 yaml 文件创建

如果不指定 `typeref` ，那么任务将在 ops-controller-manager pod 中执行。

下面是一个定时任务，每分钟检查一次 http 状态码，如果不是 200，那么发送通知。

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Task
metadata:
  name: alert-http-status-dockermirror
  namespace: ops-system
spec:
  desc: alert
  crontab: "*/1 * * * *"
  variables:
    url: http://1.1.1.1:5000/
    expect: "200"
    message: ${url} http status is not ${expect}
  steps:
    - name: get status
      content: curl -I -m 10 -o /dev/null -s -w %{http_code} ${url}
    - name: notifaction
      when: ${result} != ${expect}
      content: |
        curl -X POST 'https://xz.wps.cn/api/v1/webhook/send?key=xxx' -H 'content-type: application/json' -d '{ "msgtype": "text", "text": { "content": "${message}" } }'
```

### 查看对象

```bash
kubectl get task t1 -n ops-system

kubectl get task -n ops-system
NAME                             CRONTAB       TYPEREF   NAMEREF   NODENAME   ALL    STARTTIME   RUNSTATUS
alert-http-status-dockermirror   */1 * * * *
```
