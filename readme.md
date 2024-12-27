## Quick Start

### install Opscli

Supported OS Linux and macOS.

If Good network connections to GitHub

`curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -`

else Poor network connections to GitHub

`curl -sfL https://cf.ghproxy.cc/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh |VERSION=latest sh -`

### install Ops Controller

1. Add Helm repo

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
```

2. Install Opscli Controller

```bash
helm install myops ops/ops --version 1.0.0 --namespace ops-system --create-namespace
```

## More

Go to docs for more information.

- [中文文档](https://www.chenshaowen.com/ops/zh)
- [English Docs](https://www.chenshaowen.com/ops/en)
