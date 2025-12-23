## Quick Start

### install Opscli

Supported OS Linux and macOS.

If Good network connections to GitHub

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

else Poor network connections to GitHub

```bash
PROXY=https://ghfast.top/
curl -sfL $PROXY/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest PROXY=$PROXY sh -
```

### install Ops Controller

1. Add Helm repo

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
```

2. Install Ops Controller

```bash
helm install myops ops/ops --version 3.0.0 --namespace ops-system --create-namespace
```

## More

Go to docs for more information.

- [中文文档](https://www.chenshaowen.com/ops/zh)
- [English Docs](https://www.chenshaowen.com/ops/en)
