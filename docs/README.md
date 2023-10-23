## Ops

Ops 是一个运维工具项目。它的目标是提供一个简单的运维工具，让运维人员可以快速地完成运维工作。

## 架构

![](https://www.chenshaowen.com/blog/images/2023/04/ops-arch.png)

## 项目组成

Ops 项目包含了三个组件：

- ops-cli，一个命令行工具，辅助运维人员在命令行终端完成一些自动化的运维工作
- ops-server(完成度较低，等 Dashboard 一起开发)，一个 HTTP 服务，用于提供 HTTP API
- ops-controller，以 Operator 的形式管理主机、集群、任务等资源

## 生产实践

- 运维 10 台 32c125G 生产构建机器，近一年
- 批量运维线上集群 40+ 个集群，节点数超 400+
- 定时清理构建集群中的镜像
- 定时清理构建集群中的历史 Pod、PipelineRun、PVC 等资源
- 拨测 http 状态检测，并发送告警
- 轮询 promql 检测，并发送告警
- 集群备份
- 集群升级
- 从容器上传文件到 S3

## Todo

- Dashboard UI，准备开发一个简单的页面
- Taskhub，以 hub 的形式允许大家分享自己的 task

欢迎一起交流，关注我的公众号，发送【微信号】，即可获取到我的微信号。
