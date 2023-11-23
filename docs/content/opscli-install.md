## opscli 安装

### 安装

- 单机安装

国内使用:

```bash
curl -sfL https://mirror.ghproxy.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh |VERSION=latest sh -
```

国外使用:

```bash
curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -
```

- 批量安装

将需要安装的全部主机 ip 都写入到 `hosts.txt` 文件中，然后使用 `opscli shell` 命令批量安装，凭证默认为当前用户的 `~/.ssh/id_rsa`。

国内使用:

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://mirror.ghproxy.com/https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

国外使用:

```bash
/usr/local/bin/opscli shell --content "curl -sfL https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh | VERSION=latest sh -" -i hosts.txt
```

### 版本升级

- 单机

```bash
sudo /usr/local/bin/opscli upgrade
```

- 批量

```bash
/usr/local/bin/opscli shell --content "sudo /usr/local/bin/opscli upgrade" -i hosts.txt
```

## 自动补全

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
/usr/local/bin/opscli --help
```
