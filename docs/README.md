## Ops

Ops 是一个运维工具项目。它的目标是提供一个简单的运维工具，让运维人员可以快速的完成运维工作。

## 架构

![](https://www.chenshaowen.com/blog/images/2023/04/ops-arch.png)

## Quick Start

Ops 项目包含了三个组件：

- ops-controller， 一个 Kubernetes Operator，用于管理主机、集群、任务等资源。
- ops-server，一个 HTTP 服务，用于提供 HTTP API。
- ops-cli， 一个命令行工具，用于快速的完成运维工作。

## 生产实践

- 定时清理 Kubernetes 集群中的镜像
- 定时清理 Kubernetes 集群中的 Pod
- http 状态检测，并发送告警
- promql 检测，并发送告警
- 集群备份
- S3 文件下载上传
