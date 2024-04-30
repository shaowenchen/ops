## opscli copilot command

### 参数说明

```bash
/usr/local/bin/opscli copilot --help

use llm to assist ops

Usage:
  opscli copilot [flags]

Flags:
  -e, --endpoint string   e.g. https://api.openai.com/v1
  -h, --help              help for copilot
      --history int        (default 5)
  -k, --key string        e.g. sk-xxx
  -m, --model string      e.g. gpt-3.5-turbo
  -s, --silence
  -v, --verbose string
```

copilot 会默认从环境变量获取 `OPENAI_API_HOST`、`OPENAI_API_BASE`、`OPENAI_API_MODEL`、`OPENAI_API_KEY`、`OPS_SERVER`, `OPS_TOKEN`，如果环境变量不存在，则使用默认值。

### 使用

- 设置环境变量

```bash
export OPENAI_API_KEY=sk-xxxx
export OPENAI_API_HOST=https://llmapi.YOUR-OPENAI-SERVER.com/v1
export OPS_SERVER=http://1.1.1.1
export OPS_TOKEN=xxxx
```

- 运行 Copilot

```bash
/usr/local/bin/opscli copilot

Welcome to Opscli Copilot. Please type "exit" or "q" to quit.
Opscli>
```

- 查看支持哪些操作

```bash
Opscli> 有哪些可用的操作
这里列出了可用的操作及其相应的描述和变量：

1. list-cluster：查询K8s集群列表。
2. list-task：查询任务列表。
3. list-pipeline：查询流水线列表。
4. restart-pod：重新启动或删除Pod。变量：podname（一个或多个Pod名称）。
5. force-restart-pod：强制重新启动或删除Pod。变量：podname（一个或多个Pod名称）。
6. get-cluster-ip：查询集群的IP地址。变量：clusterip（一个或多个集群IP地址）。
7. clear-disk：清空磁盘。变量：nodeName（一个或多个节点名称）。

您可以根据具体需求选择合适的操作来执行相应的操作。
```

- 查询有哪些集群

```bash
Opscli> 有哪些集群
这些集群如下：
1. ops-system/xx-xx：该集群部署在xxx云上的88推理集群
2. ops-system/xx-xx：该集群部署在xxx上的119集群
3. ops-system/xx-xx：该集群部署在xxx上的训推一体集群
4. ops-system/xx-xx：该集群部署在xx上的 NPU 训练集群
```

- 重启一个 Pod

```bash
Opscli> 强制重启训推一体集群上的pod ubuntu-8474647969-qszcj
强制重启训推一体集群上的pod ubuntu-8474647969-qszcj

- 步骤：检查pod是否存在
- 输出：
在默认命名空间中找到了Pod ubuntu-8474647969-qszcj。

- 步骤：删除pod
- 输出：
警告：立即删除不会等待确认正在运行的资源是否已终止。该资源可能会无限期地在集群上运行。

Pod "ubuntu-8474647969-qszcj" 已被强制删除。
```

在集群上可以看到相关 Pod 的事件

```bash
kubectl get pod ubuntu-8474647969-qszcj -w
NAME                      READY   STATUS    RESTARTS   AGE
ubuntu-8474647969-qszcj   1/1     Running   0          20h
ubuntu-8474647969-qszcj   1/1     Terminating   0          20h
ubuntu-8474647969-qszcj   1/1     Terminating   0          20h
```
