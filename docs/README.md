## Ops

Ops 是一个运维工具项目。它的目标是提供一个简单的运维工具，让运维人员可以快速地完成运维工作。

## 架构

![](https://www.chenshaowen.com/blog/images/2023/04/ops-arch.png)

## 项目组成

Ops 项目包含了三个组件：

- ops-cli，一个命令行工具，辅助运维人员在命令行终端完成一些自动化的运维工作
- ops-server 一个 HTTP 服务，用于提供 HTTP API，提供有一个 Dashboard 的界面
- ops-controller，以 Operator 的形式管理主机、集群、任务等资源

## 生产实践

- 运维 10 台 32c125G 生产构建机器
- 运维线上海外集群 40+ 个集群，节点数超 400+
- 运维数千卡的 GPU 集群

## Todo

- 按照 Task 组装 Pipeline 对接场景的思路，重写 Copilot

欢迎一起交流，关注我的公众号，即可获取到我的微信号。
