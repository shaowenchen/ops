## Quick install

- Good network connections to GitHub

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

- Poor network connections to GitHub

```bash
curl -sfL https://ghproxy.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh |VERSION=latest sh -
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
  host        command about host
  kubernetes  command about kubernetes
  pipeline    command about pipeline
  storage     command about remote storage
  upgrade     upgrade to latest version
  version     get current opscli version
```

## Auto Completion

After install `bash-completion` package, run script:

```bash
echo "source /usr/share/bash-completion/bash_completion" >>~/.bashrc
echo 'source <(opscli completion bash)' >>~/.bashrc
source ~/.bashrc
```
