## Quick install

- Good network connections to GitHub

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh | VERSION=latest sh -
```

- Poor network connections to GitHub

```bash
curl -sfL https://ghproxy.com/https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh |VERSION=latest sh -
```

## Supported OS

- Linux
- macOS

## Usage

```bash
opscli --help

Usage:
  opscli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  host        config host with this command
  kube        use kubeconfig to config kubernetes
  pipeline    run pipeline with this command
  storage     config storage with this command
  upgrade     upgrade opscli version to latest
  version     get current opscli version
```

## Auto Completion

```bash
echo 'source <(opscli completion bash)' >>~/.bashrc
source ~/.bashrc
```