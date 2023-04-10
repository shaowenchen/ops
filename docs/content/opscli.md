## opscli

### 功能简介

- 创建 Ops Controller CRD 资源

主要分为三类 CRD 资源: `Host`, `Cluster`, `Task`

- 批量远程执行命令

- 批量分发文件

### 支持的操作系统

- Linux 
- macOS

## 快速安装

如果网络连接 GitHub 很好，可以使用下面的命令安装：

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

如果网络连接 GitHub 不好，可以使用下面的命令安装：

```bash
curl -sfL https://ghproxy.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh |VERSION=latest sh -
```

## 自动不全

### bash

```bash
echo 'source <(opscli completion bash)' >>~/.bashrc
```

### zsh

```bash
echo 'source <(opscli completion zsh)' >>~/.zshrc
```

## 更多

```bash
opscli --help
```

